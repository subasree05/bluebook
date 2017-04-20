package bcl

import (
	"bytes"
	"fmt"
)

type Node interface {
	Type() NodeType
	String() string
}

type NodeType int

// Type returns itself and provides an easy default implementation
// for embedding in a Node.
func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeString     NodeType = iota // one line string
	NodeIdentifier                 // an identifier
	NodeList                       // a list of nodes
	NodeBlock                      // a block defining any object in test definition
	NodeExpression                 // expression node, field = value
)

type StringNode struct {
	NodeType
	tree *Tree
	Text []byte
}

func (s *StringNode) String() string {
	return fmt.Sprintf("%q", s.Text)
}

func (t *Tree) newString(text string) *StringNode {
	return &StringNode{
		NodeType: NodeString,
		tree:     t,
		Text:     []byte(text),
	}
}

type IdentifierNode struct {
	NodeType
	tree *Tree
	Text []byte
}

func (i *IdentifierNode) String() string {
	return fmt.Sprintf("%s", i.Text)
}

func (t *Tree) newIdentifier(text string) *IdentifierNode {
	return &IdentifierNode{
		NodeType: NodeIdentifier,
		tree:     t,
		Text:     []byte(text),
	}
}

type ListNode struct {
	NodeType
	tree  *Tree
	Nodes []Node
}

func (l *ListNode) String() string {
	b := new(bytes.Buffer)
	for _, n := range l.Nodes {
		fmt.Fprint(b, n)
	}
	return b.String()
}

func (l *ListNode) append(n Node) {
	l.Nodes = append(l.Nodes, n)
}

func (t *Tree) newList() *ListNode {
	return &ListNode{
		NodeType: NodeList,
		tree:     t,
	}
}

type ExpressionNode struct {
	// expression always uses assignment operator, at least for now
	NodeType
	tree  *Tree
	Field *IdentifierNode
	Value Node
}

func (e *ExpressionNode) String() string {
	return fmt.Sprintf("%s = %s", e.Field, e.Value)
}

func (t *Tree) newExpression(field *IdentifierNode, value Node) *ExpressionNode {
	return &ExpressionNode{
		NodeType: NodeExpression,
		tree:     t,
		Field:    field,
		Value:    value,
	}
}

type BlockNode struct {
	NodeType
	tree        *Tree
	Id          *IdentifierNode   // block type, e.g. assertion or test
	Driver      *StringNode       // block driver
	Name        *StringNode       // user provided block name for referencing later
	Expressions []*ExpressionNode // list of expressions in the block
}

func (b *BlockNode) String() string {
	return fmt.Sprintf("%s %s %s { %v }",
		b.Id, b.Driver, b.Name, b.Expressions)
}

func (t *Tree) newBlock(idNode *IdentifierNode, driverNode *StringNode, nameNode *StringNode) *BlockNode {
	return &BlockNode{
		NodeType: NodeBlock,
		tree:     t,
		Id:       idNode,
		Driver:   driverNode,
		Name:     nameNode,
	}
}
