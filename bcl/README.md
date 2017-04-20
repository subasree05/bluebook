# BCL


## Grammar

```
<digit> ::= 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9
<letter*> ::= <letter> <letter*>
<letter> ::= A | B | ... | Z | a | b ... | z
<char*> ::= <char> <char*>
<char> ::= any char
<operator> ::= =
<ident> ::= <letter> <letter*>
<string> ::= " <char*> "
<comma> ::= ,
<item> ::= <string> <comma>
<item*> ::= <item> <item*>
<list> ::= [ <item*> ]
<expression> ::= <ident> <operator> <string> | <ident> <operator> <list>
<expression*> ::= <expression> <expression*>
<block> ::= <ident> <string> <string> { <expression*> }
```
