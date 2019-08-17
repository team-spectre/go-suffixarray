# go-suffixarray

Linear time suffix array generator in Go.

[![License](https://img.shields.io/github/license/team-spectre/go-suffixarray.svg?maxAge=86400)](https://github.com/team-spectre/go-suffixarray/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/team-spectre/go-suffixarray?status.svg)](https://godoc.org/github.com/team-spectre/go-suffixarray)
[![Build Status](https://img.shields.io/travis/com/team-spectre/go-suffixarray.svg?maxAge=3600&logo=travis)](https://travis-ci.com/team-spectre/go-suffixarray)
[![Issues](https://img.shields.io/github/issues/team-spectre/go-suffixarray.svg?maxAge=7200&logo=github)](https://github.com/team-spectre/go-suffixarray/issues)
[![Pull Requests](https://img.shields.io/github/issues-pr/team-spectre/go-suffixarray.svg?maxAge=7200&logo=github)](https://github.com/team-spectre/go-suffixarray/pulls)
[![Latest Release](https://img.shields.io/github/release/team-spectre/go-suffixarray.svg?maxAge=2592000&logo=github)](https://github.com/team-spectre/go-suffixarray/releases)

## Overview

[Suffix arrays][1] are a data structure that allows for _very fast searching_
of a large corpus in `O(m log n)` time, where `n` is the length of the corpus
and `m` is the length of the search string.

The basic idea for using a suffix array is that you have a search string
(`na`), a corpus (`banana`), and the sorted suffix tree for the corpus:

```
0: $       [ptr to 6]
1: a$      [ptr to 5]
2: ana$    [ptr to 3]
3: anana$  [ptr to 1]
4: banana$ [ptr to 0]
5: na$     [ptr to 4]
6: nana$   [ptr to 2]
```

We then do a binary search to locate the region where our search string, `na`,
is a prefix of the suffixes.  That's indices 5 and 6 in this case, which are
pointers to corpus offsets 4 and 2 respectively.

The actual suffix array only contains the pointers themselves, with the text
for each array slot being reconstructed on demand from the corpus and the
slot's pointer.  This implementation uses variable-length pointers, removing
the space blow-up that some other implementations have for small corpora.

## Construction

Construction of the suffix array uses the SA-IS algorithm first defined by
Nong, Zhang, and Chan in "Linear Suffix Array Construction by Almost Pure
Induced-Sorting" [link][2], using Screwtape's "A walk through the SA-IS Suffix
Array Construction Algorithm" [link][3] as a practical guide to the
implementation.

SA-IS runs in `O(n)` time.  In practice, a reasonably modern laptop can index
a hundreds-of-megabytes corpus in tens-of-minutes.

[1]: https://en.wikipedia.org/wiki/Suffix_array
[2]: https://doi.org/10.1109%2FDCC.2009.42
[3]: https://zork.net/~st/jottings/sais.html
