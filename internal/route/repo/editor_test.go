// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gogs/git-module"
	"github.com/golang/mock/gomock"
	"github.com/ivis-yoshida/gogs/internal/context"
	mock_context "github.com/ivis-yoshida/gogs/internal/mocks/context"
	mock_db "github.com/ivis-yoshida/gogs/internal/mocks/db"
	mock_repo "github.com/ivis-yoshida/gogs/internal/mocks/repo"
)

func Test_createDmp(t *testing.T) {
	// モックの呼び出しを管理するControllerを生成
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepoUtil := mock_repo.NewMockAbstructRepoUtil(mockCtrl)
	mockDmpUtil := mock_repo.NewMockAbstructDmpUtil(mockCtrl)

	mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	mockCtxRepo := mock_context.NewMockAbstructCtxRepository(mockCtrl)
	mockDbRepo := mock_db.NewMockAbstructDbRepository(mockCtrl)

	// c.CallData()された値を確認するための変数
	dummyData := make(map[string]interface{})

	tests := []struct {
		name                string
		PrepateMockDmpUtil  func() AbstructDmpUtil
		PrepareMockRepoUtil func() AbstructRepoUtil
		PrepareMockContexts func() context.AbstructContext
		wantErr             bool
	}{
		{
			name: "処理が正常に進み\"repo/editor/edit\"が描画のために呼び出される事を確認する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/orgs").Return(nil)
				mockDmpUtil.EXPECT().FetchDmpSchema(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/json_schema/schema_dmp_meti").Return(nil)

				return mockDmpUtil
			},
			PrepareMockRepoUtil: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/basic").Return([]byte(`[{"name":`), nil)

				mockRepoUtil.EXPECT().DecodeBlobContent([]byte(`[{"name":`)).Return(`[{"name":`, nil)

				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/orgs/meti").Return([]byte(`"dummySchema"}]`), nil)

				mockRepoUtil.EXPECT().DecodeBlobContent([]byte(`"dummySchema"}]`)).Return(`"dummySchema"}]`, nil)

				return mockRepoUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return("meti")
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)

				mockCtxRepo.EXPECT().GetTreePath().AnyTimes().Return("")
				mockCtxRepo.EXPECT().GetRepoLink().AnyTimes().Return("localhost:3080/dummyUser/dummyRepo")
				mockCtxRepo.EXPECT().GetBranchName().AnyTimes().Return("master")
				mockCtxRepo.EXPECT().GetDbRepo().AnyTimes().Return(mockDbRepo)
				mockCtxRepo.EXPECT().GetCommitId().AnyTimes().Return(&git.SHA1{})

				mockDbRepo.EXPECT().FullName().Return("dummyRepo")

				mockCtx.EXPECT().Success("repo/editor/edit")

				return mockCtx
			},
			wantErr: false,
		},
		{
			name: "BindingDmpSchemaListの呼び出しで失敗することを確認する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/orgs").Return(fmt.Errorf("これは想定されたエラーです"))
				return mockDmpUtil
			},
			PrepareMockRepoUtil: func() AbstructRepoUtil {
				// repoUtilのどの関数も呼び出されない想定
				return mockRepoUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return("error")
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)

				return mockCtx
			},
			wantErr: true,
		},
		{
			name: "FetchDmpSchemaで失敗することを確認する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/orgs").Return(nil)

				mockDmpUtil.EXPECT().FetchDmpSchema(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/json_schema/schema_dmp_error").Return(fmt.Errorf("これは想定されたエラーです"))

				return mockDmpUtil
			},
			PrepareMockRepoUtil: func() AbstructRepoUtil {
				// repoUtilのどの関数も呼び出されない想定
				return mockRepoUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return("error")
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)

				return mockCtx
			},
			wantErr: true,
		},
		{
			name: "１度目のFetchContentOnGithubで失敗することを確認する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/orgs").Return(nil)

				mockDmpUtil.EXPECT().FetchDmpSchema(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/json_schema/schema_dmp_meti").Return(nil)

				return mockDmpUtil
			},
			PrepareMockRepoUtil: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/basic").Return(nil, fmt.Errorf("これは想定されたエラーです"))
				return mockRepoUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return("meti")
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)

				return mockCtx
			},
			wantErr: true,
		},
		{
			name: "１度目のDecodeBlobContentで失敗することを確認する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/orgs").Return(nil)

				mockDmpUtil.EXPECT().FetchDmpSchema(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/json_schema/schema_dmp_meti").Return(nil)

				return mockDmpUtil
			},
			PrepareMockRepoUtil: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/basic").Return([]byte(`[{"name":`), nil)

				mockRepoUtil.EXPECT().DecodeBlobContent([]byte(`[{"name":`)).Return("", fmt.Errorf("これは想定されたエラーです"))
				return mockRepoUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return("meti")
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)

				return mockCtx
			},
			wantErr: true,
		},
		{
			name: "2度目のFetchContentsOnGithubで失敗することを確認する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/orgs").Return(nil)

				mockDmpUtil.EXPECT().FetchDmpSchema(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/json_schema/schema_dmp_meti").Return(nil)

				return mockDmpUtil
			},
			PrepareMockRepoUtil: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/basic").Return([]byte(`[{"name":`), nil)
				mockRepoUtil.EXPECT().DecodeBlobContent([]byte(`[{"name":`)).Return(`[{"name":`, nil)

				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/orgs/meti").Return(nil, fmt.Errorf("これは想定されたエラーです"))

				return mockRepoUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return("meti")
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)

				return mockCtx
			},
			wantErr: true,
		},
		{
			name: "2度目のDecodeBlobContentで失敗することを確認する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/orgs").Return(nil)

				mockDmpUtil.EXPECT().FetchDmpSchema(mockCtx, "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/json_schema/schema_dmp_meti").Return(nil)

				return mockDmpUtil
			},
			PrepareMockRepoUtil: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/basic").Return([]byte(`[{"name":`), nil)
				mockRepoUtil.EXPECT().DecodeBlobContent([]byte(`[{"name":`)).Return(`[{"name":`, nil)

				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/dmp/orgs/meti").Return([]byte(`"dummySchema"}]`), nil)
				mockRepoUtil.EXPECT().DecodeBlobContent([]byte(`"dummySchema"}]`)).Return("", fmt.Errorf("これは想定されたエラーです"))

				return mockRepoUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return("meti")
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)

				return mockCtx
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createDmp(tt.PrepareMockContexts(), tt.PrepareMockRepoUtil(), tt.PrepateMockDmpUtil())

			if !tt.wantErr {
				// 正常系の際はdummyDataの中身を確認
				if reflect.DeepEqual(fmt.Sprintf("%v", dummyData["FileContent"]), `{"name":"dummySchema"}`) {
					t.Errorf("DMP Schema is wrong,  got:  %v,  want: `[{\"name\":\"dummySchema\"}]`", dummyData["FileContent"])
				}
			}
		})
	}
}

func Test_fetchDmpSchema(t *testing.T) {
	// モックの呼び出しを管理するControllerを生成
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	mockRepoUtil := mock_repo.NewMockAbstructRepoUtil(mockCtrl)

	// c.CallData()された値を確認するための変数
	dummyData := make(map[string]interface{})

	tests := []struct {
		name           string
		blobPath       string
		PrepareMockFn  func() AbstructRepoUtil
		PrepareMockCtx func() context.AbstructContext
		wantErr        bool
	}{
		{
			name:     "fetchDmpSchemaの正常終了を確認する",
			blobPath: "https://right.pattern.com/dmp/orgs/meti",
			PrepareMockFn: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://right.pattern.com/dmp/orgs/meti").Return([]byte(`{"content":"SGVsbG8sIHdvcmxkLg=="}`), nil)
				mockRepoUtil.EXPECT().DecodeBlobContent([]byte(`{"content":"SGVsbG8sIHdvcmxkLg=="}`)).Return("Hello, world.", nil)
				return mockRepoUtil
			},
			PrepareMockCtx: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)
				return mockCtx
			},
			wantErr: false,
		},
		{
			name:     "FetchContentsOnGithubで失敗することを確認する",
			blobPath: "https://wrong.pattern.com/dmp/orgs/meti",
			PrepareMockFn: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://wrong.pattern.com/dmp/orgs/meti").Return(nil, fmt.Errorf("これは想定されたエラーです"))
				return mockRepoUtil
			},
			PrepareMockCtx: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().Times(0)
				return mockCtx
			},
			wantErr: true,
		},
		{
			name:     "DecodeBlobContentで失敗することを確認する",
			blobPath: "https://right.pattern.com/dmp/orgs/meti",
			PrepareMockFn: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://right.pattern.com/dmp/orgs/meti").Return([]byte(`{"data":"SGVsbG8sIHdvcmxkLg=="}`), nil)
				mockRepoUtil.EXPECT().DecodeBlobContent([]byte(`{"data":"SGVsbG8sIHdvcmxkLg=="}`)).Return("", fmt.Errorf("これは想定されたエラーです"))
				return mockRepoUtil
			},
			PrepareMockCtx: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().Times(0)
				return mockCtx
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d dmpUtil
			if err := d.fetchDmpSchema(tt.PrepareMockCtx(), tt.PrepareMockFn(), tt.blobPath); (err != nil) != tt.wantErr {
				t.Errorf("fetchDmpSchema() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if !reflect.DeepEqual(dummyData["Schema"], "Hello, world.") {
					t.Errorf("decorded Scheme is wrong,\n  got:  %v,\n  want: \"Hello, world.\"", dummyData["Schema"])
				}
			}
		})
	}
}

func Test_bidingDmpSchemaList(t *testing.T) {
	// モックの呼び出しを管理するControllerを生成
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	mockRepoUtil := mock_repo.NewMockAbstructRepoUtil(mockCtrl)

	// c.CallData()された値を確認するための変数
	dummyData := make(map[string]interface{})

	tests := []struct {
		name           string
		treePath       string
		PrepareMockFn  func() AbstructRepoUtil
		PrepareMockCtx func() context.AbstructContext
		wantErr        bool
	}{
		{
			name:     "bindingDmpSchemaListが正常終了することを確認する",
			treePath: "https://right.pattern.com/dmp/schema",
			PrepareMockFn: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://right.pattern.com/dmp/schema").Return([]byte(`[{"name":"dummySchema"}]`), nil)
				return mockRepoUtil
			},
			PrepareMockCtx: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)
				return mockCtx
			},
			wantErr: false,
		},
		{
			name:     "FetchContentsOnGithubで失敗することを確認する",
			treePath: "https://wrong.pattern.com/dmp/schema",
			PrepareMockFn: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://wrong.pattern.com/dmp/schema").Return(nil, fmt.Errorf("これは想定されたエラーです"))
				return mockRepoUtil
			},
			PrepareMockCtx: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().Times(0)
				return mockCtx
			},
			wantErr: true,
		},
		{
			name:     "Unmarshalで失敗することを確認する",
			treePath: "https://right.pattern.com/dmp/schema",
			PrepareMockFn: func() AbstructRepoUtil {
				mockRepoUtil.EXPECT().FetchContentsOnGithub("https://right.pattern.com/dmp/schema").Return([]byte(`wrong JSON`), nil)
				return mockRepoUtil
			},
			PrepareMockCtx: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().Times(0)
				return mockCtx
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d dmpUtil
			if err := d.bidingDmpSchemaList(tt.PrepareMockCtx(), tt.PrepareMockFn(), tt.treePath); (err != nil) != tt.wantErr {
				t.Errorf("bidingDmpSchemaList() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				// 正常系の際はdummyDataの中身を確認
				if !reflect.DeepEqual(dummyData["SchemaList"], []string{"dummySchema"}) {
					t.Errorf("decorded Scheme is wrong,  got:  %v,  want: []string{\"dummySchema\"}", dummyData["Schema"])
				}
			}
		})
	}
}
