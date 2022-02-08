package repo

import (
	"reflect"
	"testing"
)

func Test_fetchContentsOnGithub(t *testing.T) {
	wantByte := []byte(`{"name":"maDMP_for_test.ipynb","path":"maDMP_for_test.ipynb","sha":"859552c7e0503b939e70987e097dd2e9d236a99a","size":764,"url":"https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/maDMP_for_test.ipynb?ref=unittest","html_url":"https://github.com/ivis-kuwata/maDMP-template/blob/unittest/maDMP_for_test.ipynb","git_url":"https://api.github.com/repos/ivis-kuwata/maDMP-template/git/blobs/859552c7e0503b939e70987e097dd2e9d236a99a","download_url":"https://raw.githubusercontent.com/ivis-kuwata/maDMP-template/unittest/maDMP_for_test.ipynb","type":"file","content":"ewogImNlbGxzIjogWwogIHsKICAgImNlbGxfdHlwZSI6ICJtYXJrZG93biIs\nCiAgICJtZXRhZGF0YSI6IHt9LAogICAic291cmNlIjogWwogICAgIiMg5Y2Y\n5L2T44OG44K544OI55SobWFETVDjg4bjg7Pjg5fjg6zjg7zjg4hcbiIsCiAg\nICAiXG4iLAogICAgIuOBk+OCjOOBr+WNmOS9k+ODhuOCueODiOOBrueCuuOB\nrm1hRE1Q44OG44Oz44OX44Os44O844OI44Gn44GZ44CC44OG44K544OI57WQ\n5p6c44Gr5b2x6Z+/44KS5Y+K44G844GZ44Gf44KB44CB6Kix5Y+v44Gq44GP\n57eo6ZuG44O75YmK6Zmk44GX44Gq44GE44Gn44GP44Gg44GV44GE44CCIgog\nICBdCiAgfQogXSwKICJtZXRhZGF0YSI6IHsKICAia2VybmVsc3BlYyI6IHsK\nICAgImRpc3BsYXlfbmFtZSI6ICJQeXRob24gMyAoaXB5a2VybmVsKSIsCiAg\nICJsYW5ndWFnZSI6ICJweXRob24iLAogICAibmFtZSI6ICJweXRob24zIgog\nIH0sCiAgImxhbmd1YWdlX2luZm8iOiB7CiAgICJjb2RlbWlycm9yX21vZGUi\nOiB7CiAgICAibmFtZSI6ICJpcHl0aG9uIiwKICAgICJ2ZXJzaW9uIjogMwog\nICB9LAogICAiZmlsZV9leHRlbnNpb24iOiAiLnB5IiwKICAgIm1pbWV0eXBl\nIjogInRleHQveC1weXRob24iLAogICAibmFtZSI6ICJweXRob24iLAogICAi\nbmJjb252ZXJ0X2V4cG9ydGVyIjogInB5dGhvbiIsCiAgICJweWdtZW50c19s\nZXhlciI6ICJpcHl0aG9uMyIsCiAgICJ2ZXJzaW9uIjogIjMuOC4xMiIKICB9\nCiB9LAogIm5iZm9ybWF0IjogNCwKICJuYmZvcm1hdF9taW5vciI6IDIKfQo=\n","encoding":"base64","_links":{"self":"https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/maDMP_for_test.ipynb?ref=unittest","git":"https://api.github.com/repos/ivis-kuwata/maDMP-template/git/blobs/859552c7e0503b939e70987e097dd2e9d236a99a","html":"https://github.com/ivis-kuwata/maDMP-template/blob/unittest/maDMP_for_test.ipynb"}}`)

	type args struct {
		blobPath string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "succeed fetch blob",
			args: args{
				blobPath: "https://api.github.com/repos/ivis-kuwata/maDMP-template/contents/maDMP_for_test.ipynb?ref=unittest",
			},
			want:    wantByte,
			wantErr: false,
		},
		{
			name: "failed fetch blob",
			args: args{
				blobPath: "https://api.github.com/repos/no-exists/maDMP-template/contents/maDMP_for_test.ipynb",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchContentsOnGithub(tt.args.blobPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchContentsOnGithub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				// if !bytes.Equal(got, wantByte) {
				t.Errorf("fetchContentsOnGithub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decodeBlobContent(t *testing.T) {
	rightBlobInfo := []byte(`{"content":"SGVsbG8sIHdvcmxkLg=="}`)
	rightDecordedBlob := "Hello, world."

	wrongJsonInfo := []byte(`{"content":"SGVsbG8sIHdvcmxkLg=="`)

	type args struct {
		blobInfo []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "SucceedDecording",
			args: args{
				blobInfo: rightBlobInfo,
			},
			want:    rightDecordedBlob,
			wantErr: false,
		},
		{
			name: "FailUnmarshal",
			args: args{
				blobInfo: wrongJsonInfo,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeBlobContent(tt.args.blobInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeBlobContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("decodeBlobContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
