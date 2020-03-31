package jsonco

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

var tests = map[string]string{
	`{
		// a
		"x": "y", // b
		"x": "y", // c
	}`: `{
		    
		"x": "y",     
		"x": "y"      
	}`,
	`{
		/*
		multiline comment
		*/
		"x": "y", // b
		"x": "y", // c
	}`: `{
		  
                   
    
		"x": "y",     
		"x": "y"      
	}`,
	`{
		/*
		multiline comment with special chars in comment * * /* \* / \\ end
		*/
		"x": "y", // b
		"x": "y", // c
	}`: `{
		  
                                                                    
    
		"x": "y",     
		"x": "y"      
	}`,

	`// serve a directory
	"l/test": [
		{
		"handler": "fs",
		"dir": "../",
		// "strip_prefix": "",
		},
	],`: `                    
	"l/test": [
		{
		"handler": "fs",
		"dir": "../" 
		                      
		} 
	],`,

	`[1, 2, 3]`:                   `[1, 2, 3]`,
	`[1, 2, 3, 4,]`:               `[1, 2, 3, 4 ]`,
	`{"x":1}//[1, 2, 3, 4,]`:      `{"x":1}               `,
	`//////`:                      `      `,
	`{}/ /..`:                     `{}/ /..`,
	`{,}/ /..`:                    `{ }/ /..`,
	`{,}//..`:                     `{ }    `,
	`{[],}`:                       `{[] }`,
	`{[,}`:                        `{[ }`,
	`[[",",],]`:                   `[["," ] ]`,
	`[",\"",]`:                    `[",\"" ]`,
	`[",\"\\\",]`:                 `[",\"\\\",]`,
	`[",//"]`:                     `[",//"]`,
	`[]/* missing close at end`:   `[]                       `,
	`[]/* missing close at end *`: `[]                         `,
	`[]/* 
	missing close at end`: `[]   
                     `,
	`[",//\"
		"],`: `[",//\"
		"],`,
}

var off_test = `{
		// a
		"x": "y", /* bbb 
		*/
		"multibyte": "мультибайт"
		"x": "y", // c
	}`

var off_expect = map[int64][]int{
	12: []int{3, 2},
	58: []int{5, 22},
	78: []int{6, 2},
}

func TestMain(t *testing.T) {
	for a, b := range tests {
		buf := &bytes.Buffer{}
		io.Copy(buf, New(strings.NewReader(a)))
		a = buf.String()
		if a != b {
			a = strings.Replace(a, " ", "·", -1)
			b = strings.Replace(b, " ", "·", -1)
			t.Errorf("reader failed to clean json: \nexpected: `%s`, \n      got `%s`", b, a)
		}
	}
}

func TestOffset(t *testing.T) {
	buf := &bytes.Buffer{}
	jbuf := New(strings.NewReader(off_test))
	io.Copy(buf, jbuf)
	for off, lnc := range off_expect {
		ln, cn, err := jbuf.LineAndChar(off)
		if err != nil {
			t.Error("unexpected error", err)
		}
		if ln != lnc[0] && cn != lnc[0] {
			t.Errorf("incorrect line:char position %d:%d for offset %d: \nexpected %d:%d", ln, cn, off, lnc[0], lnc[1])
		}
	}
	_, _, err := jbuf.LineAndChar(158)
	if err == nil {
		t.Error("offset past the end of buffer did not produce an error")
	}
}
