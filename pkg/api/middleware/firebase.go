package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"projectName/pkg/api/response"
	"projectName/pkg/tlog"
	"projectName/pkg/werrors"
	"runtime"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type FirebaseAuth interface {
	MiddlewareFunc(next http.Handler) http.Handler
}

type firebaseAuth struct {
	client *auth.Client
}

type key string

const (
	AuthCtxKey key = "AUTHED_UID"
)

func (fa *firebaseAuth) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorizationヘッダーからjwtトークンを取得
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			err := errors.New("ユーザのAuthorizationが空だっためエラーとしました。")
			// どこで起きたエラーかを特定するための情報を取得
			pt, file, line, _ := runtime.Caller(0)
			funcName := runtime.FuncForPC(pt).Name()

			// エラーログ出力
			uid, ok := r.Context().Value(AuthCtxKey).(string)
			if !ok {
				tlog.GetAppLogger().Error(fmt.Sprintf("<[Unknown]Error:%+v, File: %s:%d, Function: %s>", err, file, line, funcName))
			} else {
				tlog.GetAppLogger().Error(fmt.Sprintf("<[%s]Error:%+v, File: %s:%d, Function: %s>", uid, err, file, line, funcName))
			}
			response.Error(w, r, werrors.Newf(err, http.StatusBadRequest, "認証情報がありませんでした。", "Authorization header was not found."))
			return
		}
		jwtToken := strings.Replace(authHeader, "Bearer ", "", 1)

		// JWT の検証
		authedUserToken, err := fa.client.VerifyIDToken(r.Context(), jwtToken)
		if err != nil {
			response.Error(w, r, werrors.Newf(err, http.StatusUnauthorized, "トークンが無効です", "invalid token error."))
			return
		}
		// contextにuidを格納
		r = r.WithContext(context.WithValue(r.Context(), AuthCtxKey, authedUserToken.UID))
		next.ServeHTTP(w, r)
	})
}

func CreateFirebaseInstance() FirebaseAuth {
	ctx := context.Background()

	// get credential of firebase
	var opt option.ClientOption
	gcpClient, err := secretmanager.NewClient(ctx)
	if err != nil {
		// local
		opt = option.WithCredentialsFile("firebaseCredential.json")
	} else {
		// gcp
		opt = option.WithCredentialsJSON(getFirebaseCredentialJSON(ctx, gcpClient))
	}

	// firebase appの作成
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Panic(fmt.Errorf("error initializing firebase app: %v", err))
	}

	// firebase admin clientの作成
	client, err := app.Auth(ctx)
	if err != nil {
		log.Panic(fmt.Errorf("error initialize firebase instance. %v", err))
	}

	return &firebaseAuth{
		client: client,
	}
}

// getFirebaseCredentialJSON firebaseの証明書をjsonで取得
func getFirebaseCredentialJSON(ctx context.Context, client *secretmanager.Client) []byte {
	projectID := "project-server"
	secretID := "fireauth-key"
	// requestの作成
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretID),
	}

	// get secret value
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		log.Panicf("failed to access secret version: %v", err)
	}

	return result.Payload.Data
}
