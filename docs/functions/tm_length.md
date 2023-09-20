---
title: tm_length - Functions - Configuration Language
description: The tm_length function determines the length of a collection or string.
---

# `tm_length` Function

`tm_length` determines the length of a given list, map, or string.

If given a list or map, the result is the number of elements in that collection.
If given a string, the result is the number of characters in the string.

## Examples

```
tm_length([])
0
tm_length(["a", "b"])
2
tm_length({"a" = "b"})
1
tm_length("hello")
5
```

When given a string, the result is the number of characters, rather than the
number of bytes or Unicode sequences that form them:

```
tm_length("👾🕹️")
2
```

A "character" is a _grapheme cluster_, as defined by
[Unicode Standard Annex #29](http://unicode.org/reports/tr29/). Note that
remote APIs may have a different definition of "character" for the purpose of
length limits on string arguments; a Terraform provider is responsible for
translating Terraform's string representation into that used by its respective
remote system and applying any additional validation rules to it.
