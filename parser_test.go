package gonx

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	Convey("Test Parser", t, func() {
		format := "$remote_addr [$time_local] \"$request\" $status"
		parser := NewParser(format)

		Convey("Ensure parser format is ok", func() {
			So(parser.format, ShouldEqual, format)
		})

		Convey("Test format to regexp translation", func() {
			So(parser.regexp.String(), ShouldEqual,
				`^(?P<remote_addr>[^ ]*) \[(?P<time_local>[^]]*)\] "(?P<request>[^"]*)" (?P<status>[^ ]*)$`)
		})

		Convey("ParseString", func() {
			line := `89.234.89.123 [08/Nov/2013:13:39:18 +0000] "GET /api/foo/bar HTTP/1.1" 200`
			expected := NewEntry(Fields{
				"remote_addr": "89.234.89.123",
				"time_local":  "08/Nov/2013:13:39:18 +0000",
				"request":     "GET /api/foo/bar HTTP/1.1",
				"status":      "200",
			})
			entry, err := parser.ParseString(line)
			So(err, ShouldBeNil)
			So(entry, ShouldResemble, expected)
		})

		Convey("Parse invalid string", func() {
			line := `GET /api/foo/bar HTTP/1.1`
			_, err := parser.ParseString(line)
			So(err, ShouldNotBeNil)
		})
	})
}

type ParserTestSuite struct {
	suite.Suite
	format string
	parser *Parser
}

func (suite *ParserTestSuite) SetupTest() {
	suite.format = "$remote_addr [$time_local] \"$request\" $status"
	suite.parser = NewParser(suite.format)
}

func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}

func (suite *ParserTestSuite) TestEmptyValue() {
	line := `89.234.89.123 [08/Nov/2013:13:39:18 +0000] "" 200`
	expected := NewEntry(Fields{
		"remote_addr": "89.234.89.123",
		"time_local":  "08/Nov/2013:13:39:18 +0000",
		"request":     "",
		"status":      "200",
	})
	entry, err := suite.parser.ParseString(line)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), entry, expected)
}

func TestNginxParser(t *testing.T) {
	expected := "$remote_addr - $remote_user [$time_local] \"$request\" $status \"$http_referer\" \"$http_user_agent\""
	conf := strings.NewReader(`
        http {
            include      conf/mime.types;
            log_format   minimal  '$remote_addr [$time_local] "$request"';
            log_format   main     '$remote_addr - $remote_user [$time_local] '
                                  '"$request" $status '
                                  '"$http_referer" "$http_user_agent"';
            log_format   download '$remote_addr - $remote_user [$time_local] '
                                  '"$request" $status $bytes_sent '
                                  '"$http_referer" "$http_user_agent" '
                                  '"$http_range" "$sent_http_content_range"';
        }
    `)
	parser, err := NewNginxParser(conf, "main")
	assert.NoError(t, err)
	assert.Equal(t, parser.format, expected)
}
