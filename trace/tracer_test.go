package trace

import (
	"bytes"
	"testing"
)

func NewTest(t *testing.T) {
	var buf bytes.Buffer
	tracer := New(&buf)
	if tracer == nil {
		t.Error("Newからの戻り値がnilです")
	} else {
		tracer.Trace("こんにちは, tracerパッケージ")
		if buf.String() != "こんにちは, tracerパッケージ\n" {
			t.Errorf("'%s'というあやまった文字列が出力されました", buf.String())
		}
	}
}