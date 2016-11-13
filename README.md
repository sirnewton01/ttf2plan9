ttf2plan9
==

This Go utility converts a true type font into a plan 9 font format (font and subfont)
 using the freetype port to Go.

Notes:
* Only the first 127 lower ascii characters
* No support for non-printable characters or ones that have no glyphs
* Assumes fixed width font for now
* Font output is large due to 8-bit greyscale instead of 2-bit and lack of compression

