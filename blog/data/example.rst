===================================
Example to demonstrate .rst subset
used by parser
===================================
:Author: Dr Andrew Moss
:Date: October, 2017
:Style: Page

The parser is written from scratch based on the .rst specification
located `here <docutils.sourceforge.net/docs/ref/rst/restructuredtext.html>`_.
My interpretation is that the document is re-conned from the parser
implementation in docutils. There does not appear to be a formal
specification, so I would expect divergences in behaviour from
docutils.

Restructuredtext is a general description language for documents. The
subset used here is not - these restrictions are designed to enforce
good writing style. Roughly speaking: .rst allows arbitrary nesting
of constructs and the docutils parser outputs a parse-tree of arbitrary
depth. This parser is designed to restrict the depth of the parse-tree
to ensure that the "reading context" does not grow without bound.

Motivation
==========

* Document parsers use the same underlying technology as programming
  languages - a definition as a *context free grammar*.
* They output trees of arbitrary height.
* The depth of the tree correlates to how much structure the reader
  need to **maintain understanding** of how the current piece of text
  relates to the rest of the document.
* Good writing style (e.g. Strunk & White) recommends restricted depth.
* A smaller structure corresponds to less *mental effort* for the reader
  to understand the material.

Consequences
============

1. Reserve **deep context** for the "outside" (e.g. site navigation).
2. Use a linear sequence (tree of constant bounded height) on the "inside"
   (i.e. within the document).

Design
======

Block structures
----------------

.. image:: filename

.. shell::

  Literal quoted text is indented.
  Output will be monospaced with an appropriate typographic style.

.. code::

  Following block is indented as quoted literal text.
    Indents are allowed.

  Blank lines within the block are tolerated.
  Output will be monospaced with an appropriate typographic style.
  The shell / code environment outputs will be distinct.

.. topic:: Callouts

  Indented region allows separate blocks to place inside callout box.

  1. Box continues until indent is removed.
  2. Changes in block-style are allowed

.. reference::
  :title:  Title of bilbiographic element
  :author: Author will be rendered also
  :url:    Will be converted to a link

.. quote:: An unknown poet (1327)

  The indented block will become a blockquote, attribution
  will be *added* **in an appropriate location**.

.. video:: filename

* Bullet lists are
  defined with asterisks, 
  indent continues within a single point as shown
* Subsequent bullets add to the list.

1. Only a single style of numbering is allowed.
2. Life is short.

Paragraphs typically start with the leftmost indent (other than
the specific cases shown above). The typical target does not include
heavily structured quote levels. There are inline variations of
the :code:`code style` and the :shell:`shell style`. Latex equations
can be inlined as :math:`x^n = \frac{y^n}{z^n}`.

definition
  the other form of list environment.
range
  the block up to a blank list.
indenting
  the indented part is a paragraph environment
  and so it may space multiple lines and contain
  inline styles.
