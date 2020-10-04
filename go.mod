module github.com/G-Node/gogs

go 1.14

require (
	github.com/G-Node/git-module v0.8.4-gnode
	github.com/G-Node/libgin v0.3.2
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/dustinkirkland/golang-petname v0.0.0-20191129215211-8e5a1ed0cff0
	github.com/editorconfig/editorconfig-core-go/v2 v2.3.7
	github.com/fatih/color v1.9.0 // indirect
	github.com/go-macaron/binding v1.1.1
	github.com/go-macaron/cache v0.0.0-20190810181446-10f7c57e2196
	github.com/go-macaron/captcha v0.2.0
	github.com/go-macaron/csrf v0.0.0-20190812063352-946f6d303a4c
	github.com/go-macaron/gzip v0.0.0-20160222043647-cad1c6580a07
	github.com/go-macaron/i18n v0.5.0
	github.com/go-macaron/session v0.0.0-20190805070824-1a3cdc6f5659
	github.com/go-macaron/toolbox v0.0.0-20190813233741-94defb8383c6
	github.com/gogs/chardet v0.0.0-20150115103509-2404f7772561
	github.com/gogs/cron v0.0.0-20171120032916-9f6c956d3e14
	github.com/gogs/git-module v1.1.3
	github.com/gogs/go-gogs-client v0.0.0-20200128182646-c69cb7680fd4
	github.com/gogs/go-libravatar v0.0.0-20191106065024-33a75213d0a0
	github.com/gogs/minwinsvc v0.0.0-20170301035411-95be6356811a
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/issue9/identicon v1.0.1
	github.com/jaytaylor/html2text v0.0.0-20190408195923-01ec452cbe43
	github.com/json-iterator/go v1.1.10
	github.com/klauspost/compress v1.8.6 // indirect
	github.com/klauspost/cpuid v1.2.1 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/microcosm-cc/bluemonday v1.0.4
	github.com/msteinert/pam v0.0.0-20190215180659-f29b9f28d6f9
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/niklasfasching/go-org v0.1.9
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.2.0
	github.com/prometheus/client_golang v1.6.0
	github.com/russross/blackfriday v1.5.2
	github.com/saintfish/chardet v0.0.0-20120816061221-3af4cd4741ca // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sergi/go-diff v1.1.0
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/stretchr/testify v1.6.1
	github.com/unknwon/cae v1.0.2
	github.com/unknwon/com v1.0.1
	github.com/unknwon/i18n v0.0.0-20190805065654-5c6446a380b6
	github.com/unknwon/paginater v0.0.0-20170405233947-45e5d631308e
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	golang.org/x/text v0.3.3
	gopkg.in/DATA-DOG/go-sqlmock.v2 v2.0.0-20180914054222-c19298f520d0
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/clog.v1 v1.2.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/ini.v1 v1.60.2
	gopkg.in/ldap.v2 v2.5.1
	gopkg.in/macaron.v1 v1.3.9
	gopkg.in/yaml.v2 v2.2.7
	gorm.io/driver/mysql v1.0.1
	gorm.io/driver/postgres v1.0.1
	gorm.io/driver/sqlite v1.1.3
	gorm.io/driver/sqlserver v1.0.4
	gorm.io/gorm v1.20.2
	unknwon.dev/clog/v2 v2.1.2
	xorm.io/builder v0.3.6
	xorm.io/core v0.7.2
	xorm.io/xorm v0.8.0
)

// +heroku goVersion go1.15
// +heroku install ./
