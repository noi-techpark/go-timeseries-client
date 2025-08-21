// SPDX-FileCopyrightText: NOI Techpark <digital@noi.bz.it>

// SPDX-License-Identifier: MPL-2.0

package odhts

import (
	"fmt"
	"strings"
)

type expr string

func (e expr) Build() string {
	return string(e)
}

func logicOp(op string, exs []expr) expr {
	ss := make([]string, len(exs))
	for i, v := range exs {
		ss[i] = string(v)
	}
	return expr(fmt.Sprintf("%s(%s)", op, strings.Join(ss, ",")))
}
func Quote(v string) string {
	return fmt.Sprintf("\"%s\"", v)
}
func QuoteList(l []string) []string {
	quoted := []string{}
	for _, v := range l {
		quoted = append(quoted, Quote(v))
	}
	return quoted
}

func valOp(alias string, op string, val string) expr {
	return expr(fmt.Sprintf("%s.%s.%s", alias, op, val))
}
func listOp(alias string, op string, val []string) expr {
	return expr(fmt.Sprintf("%s.%s.(%s)", alias, op, strings.Join(val, ",")))
}

// eq: Equal
func Eq(alias string, value string) expr {
	return valOp(alias, "eq", value)
}

// neq: Not Equal
func Neq(alias string, value string) expr {
	return valOp(alias, "neq", value)
}

// lt: Less Than
func Lt(alias string, value string) expr {
	return valOp(alias, "lt", value)
}

// gt: Greater Than
func Gt(alias string, value string) expr {
	return valOp(alias, "gt", value)
}

// lteq: Less Than Or Equal
func Lteq(alias string, value string) expr {
	return valOp(alias, "lteq", value)
}

// gteq: Greater Than Or Equal
func Gteq(alias string, value string) expr {
	return valOp(alias, "gteq", value)
}

// re: Regular Expression
func Re(alias string, value string) expr {
	return valOp(alias, "re", value)
}

// ire: Case Insensitive Regular Expression
func Ire(alias string, value string) expr {
	return valOp(alias, "ire", value)
}

// nre: Negated Regular Expression
func Nre(alias string, value string) expr {
	return valOp(alias, "nre", value)
}

// nire: Negated case Insensitive Regular Expression
func Nire(alias string, value string) expr {
	return valOp(alias, "nire", value)
}

// in: True if any of the values in list match
func In(alias string, values []string) expr {
	return listOp(alias, "in", values)
}

// nin: True if none of the values in list match
func Nin(alias string, values []string) expr {
	return listOp(alias, "nin", values)
}

// and: logical end between conditions (conjunction)
func And(exs []expr) expr {
	return logicOp("and", exs)
}

// and: logical or between conditions (disjunction)
func Or(exs []expr) expr {
	return logicOp("or", exs)
}

// bbi: Bounding box intersection (e.g. coordinates are at least partially within bounding box)
// SRID is optional and defaults to 4326 when left empty
func Bbi(alias string, leftX double, leftY double, rightX double, rightY double, SRID string) expr {
	if SRID == "" {
		return valOp(alias, "bbi", fmt.Sprintf("(%d,%d,%d,%d)", leftX, leftY, rightX, rightY))
	}
	return valOp(alias, "bbi", fmt.Sprintf("(%d,%d,%d,%d,%s)", leftX, leftY, rightX, rightY, SRID))
}

// bbc: Bounding box containing (e.g. coordinates are completely within bounding box)
// SRID is optional and defaults to 4326 when left empty
func Bbc(alias string, leftX double, leftY double, rightX double, rightY double, SRID string) expr {
	if SRID == "" {
		return valOp(alias, "bbc", fmt.Sprintf("(%d,%d,%d,%d)", leftX, leftY, rightX, rightY))
	}
	return valOp(alias, "bbx", fmt.Sprintf("(%d,%d,%d,%d,%s)", leftX, leftY, rightX, rightY, SRID))
}

// dlt: Distance less than (within radius of n metres around point)
// SRID is optional and defaults to 4326 when left empty
func Dlt(alias string, leftX double, leftY double, rightX double, rightY double, SRID string) expr {
	return listOp(alias, "dlt", values)
}
