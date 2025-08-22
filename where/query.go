// SPDX-FileCopyrightText: NOI Techpark <digital@noi.bz.it>

// SPDX-License-Identifier: MPL-2.0

package where

import (
	"fmt"
	"regexp"
	"strings"
)

type expr string
type ExprBuilder struct{}

func (e expr) String() string {
	return string(e)
}

func Custom(expression string) expr {
	return expr(expression)
}

func logicOp(op string, exs []expr) expr {
	ss := make([]string, len(exs))
	for i, v := range exs {
		ss[i] = v.String()
	}
	return expr(fmt.Sprintf("%s(%s)", op, strings.Join(ss, ",")))
}

func escapeSpecial(s string) string {
	// escape \,'"
	re := regexp.MustCompile(`([\\,'"])`)
	return re.ReplaceAllString(s, `\$1`)
}
func Escape(v string) string {
	return fmt.Sprintf("\"%s\"", escapeSpecial(v))
}
func EscapeList(l ...string) []string {
	quoted := []string{}
	for _, v := range l {
		quoted = append(quoted, Escape(v))
	}
	return quoted
}

func valOp(field string, op string, val string) expr {
	return expr(fmt.Sprintf("%s.%s.%s", field, op, val))
}
func listOp(field string, op string, val []string) expr {
	return expr(fmt.Sprintf("%s.%s.(%s)", field, op, strings.Join(val, ",")))
}

// eq: Equal
func Eq(field string, value string) expr {
	return valOp(field, "eq", value)
}

// neq: Not Equal
func Neq(field string, value string) expr {
	return valOp(field, "neq", value)
}

// lt: Less Than
func Lt(field string, value string) expr {
	return valOp(field, "lt", value)
}

// gt: Greater Than
func Gt(field string, value string) expr {
	return valOp(field, "gt", value)
}

// lteq: Less Than Or Equal
func Lteq(field string, value string) expr {
	return valOp(field, "lteq", value)
}

// gteq: Greater Than Or Equal
func Gteq(field string, value string) expr {
	return valOp(field, "gteq", value)
}

// re: Regular Expression
func Re(field string, value string) expr {
	return valOp(field, "re", value)
}

// ire: Case Insensitive Regular Expression
func Ire(field string, value string) expr {
	return valOp(field, "ire", value)
}

// nre: Negated Regular Expression
func Nre(field string, value string) expr {
	return valOp(field, "nre", value)
}

// nire: Negated case Insensitive Regular Expression
func Nire(field string, value string) expr {
	return valOp(field, "nire", value)
}

// in: True if any of the values in list match
func In(field string, values ...string) expr {
	return listOp(field, "in", values)
}

// nin: True if none of the values in list match
func Nin(field string, values ...string) expr {
	return listOp(field, "nin", values)
}

// and: logical end between conditions (conjunction)
func And(exs ...expr) expr {
	return logicOp("and", exs)
}

// and: logical or between conditions (disjunction)
func Or(exs ...expr) expr {
	return logicOp("or", exs)
}

// bbi: Bounding box intersection (e.g. coordinates are at least partially within bounding box)
// SRID is optional and defaults to 4326 when left empty
func Bbi(field string, lon1 float32, lat1 float32, lon2 float32, lat2 float32, SRID string) expr {
	if SRID == "" {
		return valOp(field, "bbi", fmt.Sprintf("(%v,%v,%v,%v)", lon1, lat1, lon2, lat2))
	}
	return valOp(field, "bbi", fmt.Sprintf("(%v,%v,%v,%v,%s)", lon1, lat1, lon2, lat2, SRID))
}

// bbc: Bounding box containing (e.g. coordinates are completely within bounding box)
// SRID is optional and defaults to 4326 when left empty
func Bbc(field string, lon1 float32, lat1 float32, lon2 float32, lat2 float32, SRID string) expr {
	if SRID == "" {
		return valOp(field, "bbc", fmt.Sprintf("(%v,%v,%v,%v)", lon1, lat1, lon2, lat2))
	}
	return valOp(field, "bbc", fmt.Sprintf("(%v,%v,%v,%v,%s)", lon1, lat1, lon2, lat2, SRID))
}

// dlt: Distance less than (within radius of n metres around point)
// SRID is optional and defaults to 4326 when left empty
func Dlt(field string, distM float32, lon float32, lat float32, SRID string) expr {
	if SRID == "" {
		return valOp(field, "dlt", fmt.Sprintf("(%v,%v,%v)", distM, lon, lat))
	}
	return valOp(field, "dlt", fmt.Sprintf("(%v,%v,%v,%s)", distM, lon, lat, SRID))
}

const SRID_4326 = "4326"
