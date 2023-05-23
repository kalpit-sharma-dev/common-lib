// Package command represents a structure made of property,
// operator and value.
// A Command is what a converter can work with to generate
// the Filter.
package command

import (
	"fmt"
	"strings"

	"github.com/golang-collections/collections/stack"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter"
)

//go:generate mockgen -source command.go -package command -destination ./command_mock.go

// keys to lookup patterns expected in the input filter query.
const (
	and     opKey = "And"
	or      opKey = "Or"
	gt      opKey = "Gt"
	ge      opKey = "Ge"
	lt      opKey = "Lt"
	le      opKey = "Le"
	lhs     opKey = "Lhs"
	rhs     opKey = "Rhs"
	not     opKey = "Not"
	eq      opKey = "Eq"
	like    opKey = "Like"
	null    opKey = "Null"
	nonNull opKey = "NonNull"
)

// patterns expected in the input filter query that map to a query language operator.
const (
	// And : will map to a query language AND operator.
	And Op = "AND"
	// Or : will map to a query language OR operator.
	Or Op = "OR"
	// Gt : will map to a query language GreaterThan operator.
	Gt Op = ">"
	// Ge : will map to a query language GeaterThanEqualTo operator.
	Ge Op = ">="
	// Lt : will map to a query language LessThan operator.
	Lt Op = "<"
	// Le : will map to a query language LessThanEqualTo operator.
	Le Op = "<="
	// Lhs : will map to a query language Opening Parenthesis operator.
	LHS Op = "("
	// Rhs : will map to a query language Closing Parenthesis operator.
	RHS Op = ")"
	// Not : will map to a query language NOT operator.
	Not Op = "!="
	// Eq : will map to a query language EQUALS operator.
	Eq Op = "="
	// Like : will map to a query language LIKE operator.
	Like Op = ":"
	// Null : will map to a query language IS NULL operator.
	Null Op = "IS NULL"
	// NonNull : will map to a query language IS NOT NULL operator.
	NonNull Op = "IS NOT NULL"
	// IN : will map to a query language IN operator.
	In Op = "IN"
	// Limit : will map to a query language LIMIT clause.
	Limit Op = "LIMIT"
	// OrderBy : will map to a query language ORDER BY clause.
	OrderBy Op = "ORDER BY"
)

const (
	blank string = ""

	errCannotCreateCommand   string = "cannot create command from queue : %v"
	errOperatorAlreadyExists string = "attempt to overwrite an existing operator, key: %v, value: %v with value %v"
)

var (
	errNoLHSForRHS                 error = fmt.Errorf("no LHS present for the RHS")
	errUnexpectedFilterTermination error = fmt.Errorf("did not find the expected filter termination")
)

const minWordsInCommand = 3

// Converter : interface for a command to a specific query language converter.
type Converter interface {
	DoForCommandWithoutValue(Command, func(string) string) (*filter.Filter, error)
	DoForCommandWithoutProperty(Command, func(string) string) (*filter.Filter, error)
	DoForCommandWithValue(Command, func(string) string) (*filter.Filter, error)
	AND(filters ...*filter.Filter) *filter.Filter
	OR(filters ...*filter.Filter) *filter.Filter
	GetLimitFilter(limit int) (*filter.Filter, error)
	GetOrderByFilter(field string, mapper func(string) string) (*filter.Filter, error)
}

// Op : an operator.
type Op string

// opKey : operator key.
type opKey string

// map of key vs operators.
var operators map[opKey]Op

// Command - stores a command to be run in sql.
type Command struct {
	property string
	operator string
	value    string
}

func init() {
	operators = make(map[opKey]Op)

	AddOperator(and, And)
	AddOperator(or, Or)
	AddOperator(gt, Gt)
	AddOperator(ge, Ge)
	AddOperator(lt, Lt)
	AddOperator(le, Le)
	AddOperator(lhs, LHS)
	AddOperator(rhs, RHS)
	AddOperator(not, Not)
	AddOperator(eq, Eq)
	AddOperator(null, Null)
	AddOperator(nonNull, NonNull)
	AddOperator(like, Like)
}

// AddOperator : adds a mapping of operator from filter to Command
func AddOperator(key opKey, value Op) {
	if op, ok := operators[key]; ok {
		//nolint:goerr113
		panic(fmt.Errorf(errOperatorAlreadyExists, key, op, value))
	}

	operators[key] = value
}

// New : returns a new Command.
func New(property, operator, value string) Command {
	return Command{
		property: property,
		operator: operator,
		value:    value,
	}
}

func (k opKey) equals(s string) bool {
	return s == string(operators[k])
}

// Value : returns the command value.
func (c Command) Value() string {
	return c.value
}

// Operator : returns the command operator.
func (c Command) Operator() string {
	return c.operator
}

// Property : returns the command property.
func (c Command) Property() string {
	return c.property
}

// getCommandWithValue - Returns a command from expression string.
func getCommandWithValue(queue []string) (Command, int) {
	cmd := Command{
		property: strings.TrimSpace(queue[0]),
		operator: strings.TrimSpace(queue[1]),
	}

	i := 0
	for i = 2; i < len(queue); i++ {
		cmd.value += queue[i]
		cmd.value += " "
	}

	cmd.value = strings.TrimSpace(cmd.value)

	return cmd, i
}

// getCommandWithoutValue - Returns a command from expression string.
func getCommandWithoutValue(token string) Command {
	token = strings.TrimSpace(token)

	return Command{
		operator: token,
	}
}

// Accept : accepts a visitor to run the necessary algorithm on this command.
func (c Command) Accept(converter Converter, mapper func(string) string) (*filter.Filter, error) {
	if res, err := converter.DoForCommandWithoutValue(c, mapper); res != nil {
		//nolint:wrapcheck
		return res, err
	}

	return converter.DoForCommandWithValue(c, mapper)
}

func handleBrackets(str string, brackets *stack.Stack) error {
	// Push if LHS.
	if lhs.equals(str) {
		brackets.Push(LHS)

		return nil
	}

	// Pop if RHS. Error if empty.
	if brackets.Len() == 0 {
		return errNoLHSForRHS
	}

	brackets.Pop()

	return nil
}

// GetCommandWrapperWithValidation : gets a command with validation.
func GetCommandWrapperWithValidation(words []string, brackets *stack.Stack) (*Command, int, error) {
	queue := make([]string, 0)
	spaceCounter := 0

	for _, word := range words {
		// skip processing the interim spaces in the filter.
		if word == blank {
			if len(queue) == 0 {
				return nil, 1, nil
			}

			spaceCounter++

			continue
		}

		if isEndOfCommand(word) {
			if len(queue) > 0 {
				if len(queue) < minWordsInCommand {
					//nolint:goerr113
					return nil, 0, fmt.Errorf(errCannotCreateCommand, queue)
				}

				cmd, size := getCommandWithValue(queue)

				return &cmd, size + spaceCounter, nil
			}

			// handle brackets.
			if lhs.equals(word) || rhs.equals(word) {
				err := handleBrackets(word, brackets)
				if err != nil {
					return nil, 0, err
				}
			}

			cmd := getCommandWithoutValue(word)

			return &cmd, 1, nil
		}

		queue = append(queue, word)
	}

	return nil, 0, errUnexpectedFilterTermination
}

func isEndOfCommand(word string) bool {
	if lhs.equals(word) || rhs.equals(word) || and.equals(word) || or.equals(word) {
		return true
	}

	return false
}
