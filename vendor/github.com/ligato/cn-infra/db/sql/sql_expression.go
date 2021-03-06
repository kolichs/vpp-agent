// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sql

import (
	"fmt"
)

// Expression represents part of SQL statement and optional binding ("?")
type Expression interface {
	// Stringer prints default representation of SQL to String
	// Different implementations can override this using package specific func ExpToString()
	String() string

	// Binding are values referenced ("?") from the statement
	GetBinding() []interface{}

	// Accepts calls the methods on Visitor
	Accept(Visitor)
}

// Visitor for traversing expression tree
type Visitor interface {
	VisitPrefixedExp(*PrefixedExp)
	VisitFieldExpression(*FieldExpression)
}

// PrefixedExp covers many SQL constructions. It implements sql.Expression interface.
// Instance of this structure is returned by many helper functions below.
type PrefixedExp struct {
	Prefix      string
	AfterPrefix Expression
	Suffix      string
	Binding     []interface{}
}

// String returns Prefix + " " + AfterPrefix
func (exp *PrefixedExp) String() string {
	if exp.AfterPrefix == nil {
		return exp.Prefix
	}
	return exp.Prefix + " " + exp.AfterPrefix.String()
}

// GetBinding is a getter...
func (exp *PrefixedExp) GetBinding() []interface{} {
	return exp.Binding
}

// Accept calls VisitPrefixedExp(...) & Accept(AfterPrefix)
func (exp *PrefixedExp) Accept(visitor Visitor) {
	visitor.VisitPrefixedExp(exp)
}

// FieldExpression for addressing field of an entity in SQL expression
type FieldExpression struct {
	PointerToAField interface{}
	AfterField      Expression
}

// String returns Prefix + " " + AfterPrefix
func (exp *FieldExpression) String() string {
	prefix := fmt.Sprint("<field on ", exp.PointerToAField, ">")
	if exp.AfterField == nil {
		return prefix
	}
	return prefix + " " + exp.AfterField.String()
}

// GetBinding is a getter...
func (exp *FieldExpression) GetBinding() []interface{} {
	return nil
}

// Accept calls VisitFieldExpression(...) & Accept(AfterField)
func (exp *FieldExpression) Accept(visitor Visitor) {
	visitor.VisitFieldExpression(exp)
}

// SELECT keyword of SQL expression
func SELECT(entity interface{}, afterKeyword Expression, binding ...interface{}) Expression {
	return &PrefixedExp{"SELECT", FROM(entity, afterKeyword), "", binding}
}

// FROM keyword of SQL expression
func FROM(pointerToAStruct interface{}, afterKeyword Expression) Expression {
	return &PrefixedExp{"FROM", afterKeyword, "", []interface{}{pointerToAStruct}}
}

// WHERE keyword of SQL statement
func WHERE(afterKeyword Expression) Expression {
	return &PrefixedExp{"WHERE", afterKeyword, "", nil}
}

// DELETE keyword of SQL statement
func DELETE(entity interface{}, afterKeyword Expression) Expression {
	return &PrefixedExp{"DELETE", afterKeyword, "", nil}
}

// Exp function creates instance of sql.Expression from string statement & optional binding.
// Useful for:
// - rarely used parts of SQL statements
// - create if not exists... statements
func Exp(statement string, binding ...interface{}) Expression {
	return &PrefixedExp{statement, nil, "", binding}
}

// AND keyword of SQL expression
//
// Example usage:
//
// 		WHERE(FieldEQ(&JamesBond.FirstName), AND(FieldEQ(&JamesBond.LastName)))
func AND(rigthOperand Expression) Expression {
	return &PrefixedExp{"AND", rigthOperand, "", nil}
}

// OR keyword of SQL expression
//
// Example usage:
//
// 		WHERE(FieldEQ(&PeterBond.FirstName), OR(FieldEQ(&JamesBond.FirstName)))
func OR(rigthOperand Expression) Expression {
	return &PrefixedExp{"OR", rigthOperand, "", nil}
}

// Field is a helper function to address field of a structure
//
// Example usage:
//   Where(Field(&UsersTable.LastName, UsersTable, EQ('Bond'))
//   // generates for example "WHERE last_name='Bond'"
func Field(pointerToAField interface{}, rigthOperand Expression) (exp Expression) {
	return &FieldExpression{pointerToAField, rigthOperand}
}

// FieldEQ is combination of Field & EQ on same pointerToAField
//
// Example usage:
//   FROM(JamesBond, Where(FieldEQ(&JamesBond.LastName))
//   // generates for example "WHERE last_name='Bond'"
//   // because JamesBond is a pointer to an instance of a structure that in field LastName contains "Bond"
func FieldEQ(pointerToAField interface{}) (exp Expression) {
	return &FieldExpression{pointerToAField, EQ(pointerToAField)}
}

// PK is alias FieldEQ (user for better readability)
//
// Example usage:
//   FROM(JamesBond, Where(PK(&JamesBond.LastName))
//   // generates for example "WHERE last_name='Bond'"
//   // because JamesBond is a pointer to an instance of a structure that in field LastName contains "Bond"
func PK(pointerToAField interface{}) (exp Expression) {
	return FieldEQ(pointerToAField)
}

// EQ operator "=" used in SQL expressions
func EQ(binding interface{}) (exp Expression) {
	return &PrefixedExp{"=", Exp("?", binding), "", nil}
}

// Parenthesis expression that surrounds "inside Expression" with "(" and ")"
func Parenthesis(inside Expression) (exp Expression) {
	return &PrefixedExp{"(", inside, ")", nil}
}

// IN operator of SQL expression
// 		FROM(UserTable,WHERE(FieldEQ(&UserTable.FirstName, IN(JamesBond.FirstName, PeterBond.FirstName)))
func IN(binding ...interface{}) (exp Expression) {
	return &PrefixedExp{"IN(", nil, ")", binding}
}
