# Interactive Parser in Cogent Core

* **bold at start** is not a problem now

<!--- this is a comment for sure -->  and beyond

> quote para -- what is it doing?

![figure](some text){ width=30% }

```Go
if a + b * 2 / 7 - 42 > c * d + e / 72 {
  for a := 0; a < 10; a++ {
    switch a {
    case 1:
    }
  }
}
```

The `parse` *package* __allows__ regular **users** to **create** things __parsers *and nested* __ using _the_ Cogent Core graphical interface system.  [link tag](and addresses) are parsed as expected.  Including plain [tags] without any addresses.  https://wikipedia.com should work too once we plug in url recognizer.

This is not a # heading

## SubHead

### SubSubHead

- [x] done item
- [ ] not done item

* Top level item
    + sub item
    - sub item
    * sub item
	

```Go
This is a quoted item -- we will grab the language and switch parsers!
```

Parse uses a robust, top-down __Recursive Descent (RD)__ parsing technique (see [WikiPedia](https://en.wikipedia.org/wiki/Recursive_descent_parser)), which is the approach used by most hand-coded parsers, which are by far the most widely used in practice (e.g., for **gcc**, **clang**, and **Go**) for [various reasons](http://blog.reverberate.org/2013/09/ll-and-lr-in-context-why-parsing-tools.html) -- see this [stack overflow](https://stackoverflow.com/questions/6319086/are-gcc-and-clang-parsers-really handwritten) thread too.  As far as we can tell (e.g., from this list on [WikiPedia](https://en.wikipedia.org/wiki/Comparison_of_parser_generators) ) there are not many recursive-descent *parser generators*, and none that use the same robust, fast techniques that we employ in parse.

It seems that most parsing is dominated by a very strong *sequentiality assumption* -- that you must parse everything in a strictly sequential, incremental, left-to-right, one-token-at-a-time manner.  If you step outside of that box (or break with the [herd](https://en.wikipedia.org/wiki/GNU_Bison) if you will), by loading the entire source in to RAM and processing the entire thing as a whole structured entity (which is entirely trivial these days -- even the biggest source code is typically tiny relative to RAM capacity), then much simpler, more robust solutions are possible.  In other words, instead of using a "1D" solution, we actually use a **3D** solution to parsing (line, char, and nesting depth).  Specifically, we just carve the whole source in to statement-level chunks and then proceed to look for distinctive **Key Tokens** *anywhere* in the statement to determine what kind of statement it is, and then proceed recursively to carve that up into its respective parts, using the same approach.  There is never any backtracking or shift-reduce conflicts or any of those annoying issues that plague other approaches.  And no need for complicated lookup tables etc (e.g., as needed for the `LL(k)` approach used in the popular [ANTLR](https://en.wikipedia.org/wiki/ANTLR) parser).  The whole process is very fast, simple, and transparent.

Specifically, there are three distinct passes through the source file, each creating a solid foundation upon which the next step operates, producing significantly faster, simpler, and more error-tolerant (robust) results overall.

* **Lexer** -- takes the raw text and turns it into lexical `Tokens` that categorize a sequence of characters as things like a `Name` (identifier -- a string of letters and numbers without quotes) or a `Literal` of some type, for example a `LitStr` string that has some kind of quotes around it, or a `LitNum` which is a number.  There is a nice category (Cat) and subcategory (SubCat) level of organization to these tokens (see `token/token.go`).  Comments are absorbed in this step as well.  The key advantage for subsequent steps is that any ambiguity about e.g., syntactic elements showing up in comments or strings is completely eliminated right at the start.  Furthermore, the tokenized version of the file is much, much more compact and contains only the essential information for parsing.

* **StepTwo** -- this is a critical second pass through the lexical tokens, performing two important things:

    + **Nesting Depth** -- all programming languages use some form of parens `( )` brackets `[ ]` and braces `{ }` to group elements, and parsing must be sensitive to these.  Instead of dealing with these issues locally at every step, we do a single pass through the entire tokenized version of the source and compute the depth of every token.  Then, the token matching in parsing only needs to compare relative depth values, without having to constantly re-compute that.
	
    + **EOS Detection** -- This step detects *end of statement* tokens, which provide an essential first-pass rough-cut chunking of the source into *statements*.  In C / C++ / Go these are the *semicolons* `;` (in Go, semicolons are mostly automatically computed from tokens that appear at the end of lines -- Pi supports this as well).  In Python, this is the end of line itself, unless it is not at the same nesting depth as at the start of the line.
	
* **Parsing** -- finally we parse the tokenized source using rules that match patterns of tokens.  Instead of using bottom-up local token-sequence driven parsing, as is done in tools like `yacc` and `bison`, we use the much more robust recursive descent technique, which starts in our case with those rough-cut statement chunks produced in StepTwo, and proceeds by recognizing the type of each statement, and then further carving each one into its appropriate sub-parts, again in a top-down, progressively narrowing fashion, until everything has been categorized to the lowest level.  At each step, nodes in an __Abstract Syntax Tree (AST)__ are created, representing this same top-down broad-to-narrow parsing of the source.  Thus, the highest nodes are statement-level nodes, each of which then contain the organized elements of each statement.  These nodes are all in the natural *functional* hierarchical ordering, *not* in the raw left-to-right order of the source, and directly correspond to the way that the parsing proceeds.  Thus, building the AST at the same time as parsing is very natural in the top-down RD framework, unlike traditional bottom-up approaches, and is a major reason that hand-coded parsers use this technique.

## RD Parsing Advantages and Issues

Our top-down approach is generally much more robust: instead of depending on precise matches at every step along the way, it starts with the "big picture" by looking for the *Key Token* that unambiguously indicates what kind of thing we're looking at in a given region, and it then proceeds to process that region as such.  If some parts of that region happen to be malformed, it can issue appropriate errors, *but that will not spill over into processing of the next region.*  Thus, errors are automatically "sandboxed" in these regions, and do not accumulate.   By contrast, in bottom-up parsers, you need to add extensive error-matching rules at every step to achieve this kind of robustness, and that is often a tricky trial-and-error process and is not inherently robust.

### Solving the Associativity problem with RD parsing: Put it in Reverse!

One major problem with RD parsing is that it gets the [associativity](https://en.wikipedia.org/wiki/Operator_associativity) of mathematical operators [backwards](https://eli.thegreenplace.net/2009/03/14/some-problems-of-recursive-descent-parsers/).  To solve this problem, we simply run those rules in reverse: they scan their region from right to left instead of left to right.  This is much simpler than other approaches and works perfectly -- and is again something that you wouldn't even consider from the standard sequential mindset.  You just have to add a '-' minus sign at the start of the `Rule` to set the rule to run in reverse -- this must be set for all binary mathematical operators (e.g., `BinaryExpr` in the standard grammar).  

Also, for RD parsing, to deal properly with the [order of operations](https://en.wikipedia.org/wiki/Order_of_operations), you have to order the rules in the *reverse* order of precedence.  Thus, it matches the *lowest* priority items first, and those become the "outer branch" of the AST, which then proceeds to fill in so that the highest-priority items are in the "leaves" of the tree, which are what gets processed first.

## Lexing Rules

* order matters! and state matters!

## Parsing Rules

* order matters!

### Generative Expression Subdomains

There are certain subdomains that have very open-ended combinatorial "generative" expressive power.  These are particular challenges for any parser, and there are a few critical issues and tips for the Pi parser.

#### Arithmetic with Binary and Unary Operators

You can create arbitrarily long expressions by stringing together sequences of binary and unary mathematical / logical operators.  From the top-down parser's perspective, here are the key points:

1. Each operator must be uniquely recognizable from the soup of tokens, and this critically includes distinguishing unary from binary: e.g., correctly recognizing the binary and unary - signs here: `a - b * -c`  

2. The operators must be organized in *reverse* order of priority, so that the lowest priority operations are factored out first, creating the highest-level, broadest splits of the overall expression (in the AST tree), and then progressively finer, tighter, inner steps are parsed out.  Thus, for example in this expression:

```Go
if a + b * 2 / 7 - 42 > c * d + e / 72
```

The broadest, first split is into the two sides of the `>` operator, and then each of those sides is progressively organized first into an addition operator, then the `*` and `/`.  

3. The binary operators provide the recursive generativity for the expression.  E.g., Addition is specified as:

```Go
AddExpr: Expr '+' Expr
```

so it just finds the + token and then descends recursively to unpack each of those `Expr` chunks on either side, until there are no more tokens left there.

One particularly vexing situation arises if you have the possibility of mixing multiplication with de-pointering, both of which are indicated by the `*` symbol.  In Go, this is particularly challenging because of the frequent use of type literals, including those with pointer types, in general expressions.  We had to add the ability to specify an "exclusion" expression that detects this alternate use and excludes it for the general multiplication case (because multiplication is broader than de-pointer and thus comes first in the matching hierarchy).

#### Path-wise Operators

Another generative domain are the path-wise operators, principally the "selector" operator `.` and the slice operator `'[' SliceExpr ']'`, which can be combined with method calls and other kinds of primary expressions in a very open-ended way, e.g.,:

```Go
ps.Errs[len(ps.Errs)-1].Error()[0][1].String()
```

In the top-down parser, it is essential to create this open-ended scenario by including pre-and-post expressions surrounding the `Slice` and `Selector` operators, which then act like the Expr groups surrounding the AddExpr operator to support recursive chaining.  For Selector, the two Expr's are required, but for Slice, they are optional - that works fine:

```Go
Slice: ?PrimaryExpr '[' SliceExpr ']' ?PrimaryExpr
```

Without those optional exprs on either side, the top-down parser would just stop after getting either side of that expression.

As with the arithmetic case, order matters and in the same inverse way, where you want to match the broader items first.  Also, those cases that are more specific and would otherwise be absorbed by a more general expression (e.g., 


