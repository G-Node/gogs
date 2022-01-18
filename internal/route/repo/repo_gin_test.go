package repo

import (
	"reflect"
	"testing"
)

func Test_fetchBlobOnGithub(t *testing.T) {
	wantByte := []byte(`{"name":"maDMP_for_test.ipynb","path":"maDMP_for_test.ipynb","sha":"41c937ff8c8c14ffb53ba3d58a6f9ca30db96160","size":2112,"url":"https://api.github.com/repos/ivis-kuwata/maDMP-test/contents/maDMP_for_test.ipynb?ref=main","html_url":"https://github.com/ivis-kuwata/maDMP-test/blob/main/maDMP_for_test.ipynb","git_url":"https://api.github.com/repos/ivis-kuwata/maDMP-test/git/blobs/41c937ff8c8c14ffb53ba3d58a6f9ca30db96160","download_url":"https://raw.githubusercontent.com/ivis-kuwata/maDMP-test/main/maDMP_for_test.ipynb","type":"file","content":"ewogImNlbGxzIjogWwogIHsKICAgImNlbGxfdHlwZSI6ICJtYXJrZG93biIs\nCiAgICJtZXRhZGF0YSI6IHt9LAogICAic291cmNlIjogWwogICAgIiMg44OG\n44K544OI55So44OH44O844K/XG4iLAogICAgIlxuIiwKICAgICLjgZPjgozj\nga9HT0dT44GuaW50ZXJuYWwvcm91dGUvcmVwby9yZXBvX2dpbl90ZXN0Lmdv\n44GM5Yip55So44GZ44KL44OH44O844K/44Gn44GZ44CC6Kix5Y+v44Gq44GP\n57eo6ZuG44GX44Gq44GE44Gn44GP44Gg44GV44GE44CCIgogICBdCiAgfSwK\nICB7CiAgICJjZWxsX3R5cGUiOiAiY29kZSIsCiAgICJleGVjdXRpb25fY291\nbnQiOiBudWxsLAogICAibWV0YWRhdGEiOiB7CiAgICAic2Nyb2xsZWQiOiB0\ncnVlCiAgIH0sCiAgICJvdXRwdXRzIjogW10sCiAgICJzb3VyY2UiOiBbCiAg\nICAiIyBETVDmg4XloLFcbiIsCiAgICAiZmllbGQgPSAnJXYnIgogICBdCiAg\nfSwKICB7CiAgICJjZWxsX3R5cGUiOiAiY29kZSIsCiAgICJleGVjdXRpb25f\nY291bnQiOiBudWxsLAogICAibWV0YWRhdGEiOiB7CiAgICAic2Nyb2xsZWQi\nOiB0cnVlCiAgIH0sCiAgICJvdXRwdXRzIjogW10sCiAgICJzb3VyY2UiOiBb\nCiAgICAiIyDjg6/jg7zjgq/jg5Xjg63jg7zjg4bjg7Pjg5fjg6zjg7zjg4jl\nj5blvpdcbiIsCiAgICAiJXNoXG4iLAogICAgImdpdCBjbG9uZSBodHRwczov\nL2dpdGh1Yi5jb20vaXZpcy1rdXdhdGEvd29ya2Zsb3ctdGVtcGxhdGUuZ2l0\nIC9ob21lL2pvdnlhbi9XT1JLRkxPV1xuIiwKICAgICJybSAtciAvaG9tZS9q\nb3Z5YW4vV09SS0ZMT1cvLmdpdCIKICAgXQogIH0sCiAgewogICAiY2VsbF90\neXBlIjogImNvZGUiLAogICAiZXhlY3V0aW9uX2NvdW50IjogbnVsbCwKICAg\nIm1ldGFkYXRhIjogewogICAgInNjcm9sbGVkIjogdHJ1ZQogICB9LAogICAi\nb3V0cHV0cyI6IFtdLAogICAic291cmNlIjogWwogICAgIiMgZG1wLmpzb27j\ngatcImZpZWxkc1wi44OX44Ot44OR44OG44Kj44GM44GC44KL5oOz5a6aXG4i\nLAogICAgImltcG9ydCBvc1xuIiwKICAgICJpbXBvcnQgZ2xvYlxuIiwKICAg\nICJpbXBvcnQgc2h1dGlsXG4iLAogICAgIlxuIiwKICAgICIjIHBhdGhfZmxv\nd3MgPSBvcy5wYXRoLmpvaW4oJ1dPUktGTE9XJywgJ0ZMT1cnKVxuIiwKICAg\nICJ0bXBfcGF0aCA9ICdGTE9XJyAjIG1hRE1Q5qSc6Ki855So44OR44K5XG4i\nLAogICAgIlxuIiwKICAgICIjIOaknOiovOOBjOe1guOCj+OBo+OBn+OCiXRt\ncF9wYXRo44KScGF0aF9mbG93c+OBq+S/ruato+OBruOBk+OBqFxuIiwKICAg\nICJ0ZW1wbGF0ZXMgPSBnbG9iLmdsb2Iob3MucGF0aC5qb2luKHRtcF9wYXRo\nLCAnKionKSwgcmVjdXJzaXZlPVRydWUpXG4iLAogICAgIlxuIiwKICAgICIj\nIOmBuOaKnuWkluOBruWIhumHjuOBruOCu+OCr+OCt+ODp+ODs+e+pOOCkuWJ\niumZpFxuIiwKICAgICJmb3IgdG1wbCBpbiB0ZW1wbGF0ZXM6XG4iLAogICAg\nIiAgICBmaWxlID0gb3MucGF0aC5iYXNlbmFtZSh0bXBsKVxuIiwKICAgICIg\nICAgaWYgbm90IG9zLnBhdGguaXNkaXIodG1wbCkgYW5kIG9zLnBhdGguc3Bs\naXRleHQoZmlsZSlbMV0gPT0gJy5pcHluYic6XG4iLAogICAgIiAgICAgICAg\naWYgJ2Jhc2VfJyBub3QgaW4gZmlsZSBhbmQgZmllbGQgbm90IGluIGZpbGU6\nXG4iLAogICAgIiAgICAgICAgICAgIG9zLnJlbW92ZSh0bXBsKVxuIgogICBd\nCiAgfSwKIF0sCiAibWV0YWRhdGEiOiB7CiAgImtlcm5lbHNwZWMiOiB7CiAg\nICJkaXNwbGF5X25hbWUiOiAiUHl0aG9uIDMgKGlweWtlcm5lbCkiLAogICAi\nbGFuZ3VhZ2UiOiAicHl0aG9uIiwKICAgIm5hbWUiOiAicHl0aG9uMyIKICB9\nLAogICJsYW5ndWFnZV9pbmZvIjogewogICAiY29kZW1pcnJvcl9tb2RlIjog\newogICAgIm5hbWUiOiAiaXB5dGhvbiIsCiAgICAidmVyc2lvbiI6IDMKICAg\nfSwKICAgImZpbGVfZXh0ZW5zaW9uIjogIi5weSIsCiAgICJtaW1ldHlwZSI6\nICJ0ZXh0L3gtcHl0aG9uIiwKICAgIm5hbWUiOiAicHl0aG9uIiwKICAgIm5i\nY29udmVydF9leHBvcnRlciI6ICJweXRob24iLAogICAicHlnbWVudHNfbGV4\nZXIiOiAiaXB5dGhvbjMiLAogICAidmVyc2lvbiI6ICIzLjcuMTAiCiAgfQog\nfSwKICJuYmZvcm1hdCI6IDQsCiAibmJmb3JtYXRfbWlub3IiOiAyCn0K\n","encoding":"base64","_links":{"self":"https://api.github.com/repos/ivis-kuwata/maDMP-test/contents/maDMP_for_test.ipynb?ref=main","git":"https://api.github.com/repos/ivis-kuwata/maDMP-test/git/blobs/41c937ff8c8c14ffb53ba3d58a6f9ca30db96160","html":"https://github.com/ivis-kuwata/maDMP-test/blob/main/maDMP_for_test.ipynb"}}`)

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
				blobPath: "https://api.github.com/repos/ivis-kuwata/maDMP-test/contents/maDMP_for_test.ipynb",
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
			got, err := fetchBlobOnGithub(tt.args.blobPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchBlobOnGithub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fetchBlobOnGithub() = %v, want %v", got, tt.want)
			}
		})
	}
}
