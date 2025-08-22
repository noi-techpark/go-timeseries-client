// SPDX-FileCopyrightText: NOI Techpark <digital@noi.bz.it>

// SPDX-License-Identifier: MPL-2.0

package where

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestAnd(t *testing.T) {
	actual := And(
		Or(Eq("scode", "123"),
			Neq("scode", "456"),
			Lt("mvalue", "1"),
			Gt("mvalue", "1"),
			Lteq("mvalue", "1"),
			Gteq("mvalue", "1"),
			Re("sname", Escape("test.*")),
			Ire("sname", Escape("test.*")),
			Nre("sname", Escape("test.*")),
			Nire("sname", Escape(`test,'"\`)),
		),
		In("mvalue", "1", "2", "3"),
		Nin("mvalue", EscapeList("1", "2", "3")...),
		Bbi("scoordinate", 11.3, 46.4, 12, 47, ""),
		Bbc("scoordinate", 11.3, 46.4567567, 12, 47, SRID_4326),
		Dlt("scoordinate", 4000, 11.2, 46.7, SRID_4326),
	)

	expect := `and(or(`
	expect += `scode.eq.123,scode.neq.456,mvalue.lt.1,mvalue.gt.1,mvalue.lteq.1,mvalue.gteq.1,`
	expect += `sname.re."test.*",sname.ire."test.*",sname.nre."test.*",sname.nire."test\,\'\"\\"),`
	expect += `mvalue.in.(1,2,3),mvalue.nin.("1","2","3"),`
	expect += `scoordinate.bbi.(11.3,46.4,12,47),scoordinate.bbc.(11.3,46.456757,12,47,4326),scoordinate.dlt.(4000,11.2,46.7,4326)`
	expect += `)`

	if actual.String() != expect {
		t.Log("query did not match. dumping actual and expected..")
		t.Log(actual)
		t.Log(expect)
		t.Fail()
	}
}

func TestEscapeSpecial(t *testing.T) {
	actual := escapeSpecial(`test,test"test'test\`)
	expected := `test\,test\"test\'test\\`
	assert.Equal(t, actual, expected)
}
