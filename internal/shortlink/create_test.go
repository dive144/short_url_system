package shortlink

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 用模拟数据库验证真实的 Handler -> Service -> Repository 调用，
// 不连接本机 MySQL，也不会修改用户的数据。
func newTestRepository(t *testing.T) (*ShortLinkRepository, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		mock.ExpectClose()
		if err := sqlDB.Close(); err != nil {
			t.Error(err)
		}
	})
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewShortLinkRepository(db, nil), mock
}

func expectCreate(mock sqlmock.Sqlmock, insertErr error) {
	mock.ExpectBegin()
	insert := mock.ExpectExec("INSERT INTO `short_urls`")
	if insertErr != nil {
		insert.WillReturnError(insertErr)
		mock.ExpectRollback()
	} else {
		insert.WillReturnResult(sqlmock.NewResult(42, 1))
		mock.ExpectCommit()
	}
}

func expectMissingCode(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT .* FROM `short_urls`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "short_code", "original_url"}))
}

func TestCreateShortLinkRejectsInvalidParametersBeforeDatabase(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	now := time.Now()
	cases := []struct {
		name    string
		request *CreateShortUrlRequest
		wantErr error
	}{
		{"nil request", nil, ErrInvalidOriginalURL},
		{"empty URL", &CreateShortUrlRequest{}, ErrInvalidOriginalURL},
		{"spaces only", &CreateShortUrlRequest{OriginalUrl: "   "}, ErrInvalidOriginalURL},
		{"relative URL", &CreateShortUrlRequest{OriginalUrl: "/page"}, ErrInvalidOriginalURL},
		{"no scheme", &CreateShortUrlRequest{OriginalUrl: "example.com"}, ErrInvalidOriginalURL},
		{"unsupported scheme", &CreateShortUrlRequest{OriginalUrl: "ftp://example.com"}, ErrInvalidOriginalURL},
		{"javascript scheme", &CreateShortUrlRequest{OriginalUrl: "javascript:alert(1)"}, ErrInvalidOriginalURL},
		{"missing host", &CreateShortUrlRequest{OriginalUrl: "https:///page"}, ErrInvalidOriginalURL},
		{"port without host", &CreateShortUrlRequest{OriginalUrl: "http://:8080"}, ErrInvalidOriginalURL},
		{"malformed escape", &CreateShortUrlRequest{OriginalUrl: "https://example.com/%zz"}, ErrInvalidOriginalURL},
		{"space in host", &CreateShortUrlRequest{OriginalUrl: "https://exa mple.com"}, ErrInvalidOriginalURL},
		{"past expiry", &CreateShortUrlRequest{OriginalUrl: "https://example.com", ExpireAt: &past}, ErrInvalidExpireTime},
		{"expiry not in future", &CreateShortUrlRequest{OriginalUrl: "https://example.com", ExpireAt: &now}, ErrInvalidExpireTime},
		{"long code", &CreateShortUrlRequest{OriginalUrl: "https://example.com", CustomUrl: "abcdefghi"}, ErrCustomShortLink},
		{"slash in code", &CreateShortUrlRequest{OriginalUrl: "https://example.com", CustomUrl: "a/b"}, ErrCustomShortLink},
		{"non ASCII code", &CreateShortUrlRequest{OriginalUrl: "https://example.com", CustomUrl: "短码"}, ErrCustomShortLink},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// nil Repository 确保参数校验无需访问数据库。
			service := NewShortLinkService(nil)
			result, err := service.CreateShortLink(context.Background(), tc.request)
			if result != nil || !errors.Is(err, tc.wantErr) {
				t.Fatalf("result=%v error=%v, want %v", result, err, tc.wantErr)
			}
		})
	}
}

func TestCreateShortLinkAcceptsValidParameters(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	cases := []struct {
		name string
		req  CreateShortUrlRequest
	}{
		{"permanent", CreateShortUrlRequest{OriginalUrl: "https://example.com"}},
		{"future expiry", CreateShortUrlRequest{OriginalUrl: "https://example.com", ExpireAt: &future}},
		{"trim URL", CreateShortUrlRequest{OriginalUrl: "  https://example.com/path?q=1#section  "}},
		{"local URL", CreateShortUrlRequest{OriginalUrl: "http://localhost:9090/page"}},
		{"IPv6 URL", CreateShortUrlRequest{OriginalUrl: "http://[::1]:9090/page"}},
		{"eight character code", CreateShortUrlRequest{OriginalUrl: "https://example.com", CustomUrl: "Abcd1234"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newTestRepository(t)
			if tc.req.CustomUrl != "" {
				expectMissingCode(mock)
			}
			expectCreate(mock, nil)
			wantURL := strings.TrimSpace(tc.req.OriginalUrl)
			result, err := NewShortLinkService(repo).CreateShortLink(context.Background(), &tc.req)
			if err != nil {
				t.Fatal(err)
			}
			if result.ID != 42 || result.OriginalURL != wantURL || result.ExpireAt != tc.req.ExpireAt {
				t.Fatalf("unexpected result: %+v", result)
			}
			if tc.req.CustomUrl != "" && result.ShortCode != tc.req.CustomUrl {
				t.Fatalf("custom code changed: %q", result.ShortCode)
			}
			if tc.req.CustomUrl == "" && len(result.ShortCode) != 6 {
				t.Fatalf("invalid generated code: %q", result.ShortCode)
			}
		})
	}
}

func TestCreateShortLinkHandlerResponses(t *testing.T) {
	valid := `{"original_url":"https://example.com"}`
	custom := `{"original_url":"https://example.com","custom_code":"abcd"}`
	internalErr := errors.New("database private-host:3306 password=private-value")
	cases := []struct {
		name   string
		body   string
		setup  func(sqlmock.Sqlmock)
		status int
	}{
		{"success", valid, func(m sqlmock.Sqlmock) { expectCreate(m, nil) }, 201},
		{"invalid JSON", `{`, nil, 400},
		{"missing URL", `{}`, nil, 400},
		{"invalid URL", `{"original_url":"abc"}`, nil, 400},
		{"invalid expiry format", `{"original_url":"https://example.com","expire_at":"bad"}`, nil, 400},
		{"past expiry", `{"original_url":"https://example.com","expire_at":"2000-01-01T00:00:00Z"}`, nil, 400},
		{"invalid short code", `{"original_url":"https://example.com","custom_code":"a/b"}`, nil, 400},
		{"short code too long", `{"original_url":"https://example.com","custom_code":"abcdefghi"}`, nil, 400},
		{"null expiry is permanent", `{"original_url":"https://example.com","expire_at":null}`,
			func(m sqlmock.Sqlmock) { expectCreate(m, nil) }, 201},
		{"eight character code accepted", `{"original_url":"https://example.com","custom_code":"Abcd1234"}`,
			func(m sqlmock.Sqlmock) { expectMissingCode(m); expectCreate(m, nil) }, 201},
		{"existing custom code", custom, func(m sqlmock.Sqlmock) {
			m.ExpectQuery("SELECT .* FROM `short_urls`").WillReturnRows(
				sqlmock.NewRows([]string{"id", "short_code", "original_url"}).AddRow(1, "abcd", "https://example.com"))
		}, 409},
		{"concurrent custom code conflict", custom, func(m sqlmock.Sqlmock) {
			// 预查询时不存在，但插入时已被另一个请求抢占。
			expectMissingCode(m)
			expectCreate(m, &mysqldriver.MySQLError{Number: 1062, Message: "Duplicate entry"})
		}, 409},
		{"lookup failure", custom, func(m sqlmock.Sqlmock) {
			m.ExpectQuery("SELECT .* FROM `short_urls`").WillReturnError(internalErr)
		}, 500},
		{"insert failure", valid, func(m sqlmock.Sqlmock) { expectCreate(m, internalErr) }, 500},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newTestRepository(t)
			if tc.setup != nil {
				tc.setup(mock)
			}
			handler := NewShortLinkHandler(NewShortLinkService(repo))
			router := gin.New()
			router.POST("/api/v1/shorten", handler.CreateShortLink)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			var body struct {
				Code     int    `json:"code"`
				Message  string `json:"message"`
				ShortURL string `json:"ShortUrl"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if response.Code != tc.status || body.Code != tc.status {
				t.Fatalf("HTTP=%d body=%s, want status=%d", response.Code, response.Body.String(), tc.status)
			}
			if tc.status == 201 && body.ShortURL == "" {
				t.Fatal("missing short code in successful response")
			}
			if tc.status == 500 && body.Message != "服务器内部错误" {
				t.Fatalf("internal error leaked: %s", response.Body.String())
			}
		})
	}
}

func TestRepositoryCreateMapsDuplicateErrors(t *testing.T) {
	duplicate := &mysqldriver.MySQLError{Number: 1062, Message: "Duplicate entry"}
	otherErr := &mysqldriver.MySQLError{Number: 1045, Message: "Access denied"}
	cases := []struct {
		name    string
		dbErr   error
		wantErr error
	}{
		{"success", nil, nil},
		{"MySQL duplicate", duplicate, ErrShortCodeConflict},
		{"wrapped duplicate", fmt.Errorf("driver: %w", duplicate), ErrShortCodeConflict},
		{"GORM translated duplicate", gorm.ErrDuplicatedKey, ErrShortCodeConflict},
		{"other error", otherErr, otherErr},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newTestRepository(t)
			expectCreate(mock, tc.dbErr)
			err := repo.CreateShortLink(context.Background(), &ShortUrl{
				ShortCode: "abcd", OriginalURL: "https://example.com",
			})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error=%v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestRedirectExpiredLinkStillReturnsGone(t *testing.T) {
	repo, mock := newTestRepository(t)
	mock.ExpectQuery("SELECT .* FROM `short_urls`").WillReturnRows(
		sqlmock.NewRows([]string{"id", "short_code", "original_url", "expire_at"}).
			AddRow(1, "abcd", "https://example.com", time.Now().Add(-time.Hour)))
	handler := NewShortLinkHandler(NewShortLinkService(repo))
	router := gin.New()
	router.GET("/shortlink/:shorturl", handler.Redirect)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/shortlink/abcd", nil))
	if response.Code != http.StatusGone {
		t.Fatalf("HTTP=%d body=%s, want 410", response.Code, response.Body.String())
	}
}
