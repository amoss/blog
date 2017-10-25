============================================
Part I: Lua
============================================
:Author: Dr Andrew Moss
:Date: 2017 LP4
:CourseCode: DV1564
:CourseName: Interpreters and Scripting
:Style: Slides


+-----------------------------------+------------------------------------------+
+ Part I                            + Language Design                          |
+-----------------------------------+------------------------------------------+
+                                   + Introduction to Lua                      |
+-----------------------------------+------------------------------------------+
+                                   + C API                                    |
+-----------------------------------+------------------------------------------+
+ Reading Advice                    + To be added                              |
+-----------------------------------+------------------------------------------+


Interpreters
------------

.. topic:: Interpreter (common definitions)

  1. A program that can analyse and execute a program line by line [Google].
  2. A computer program that changes the instructions in another program into a form that can be easily understood by a computer [Cambridge].
  3. A computer program that executes each of a set of high-level instructions before going to the next instruction [Mirriam-Webster].

What is an interpreter? #Cols
-----------------------------

.. topic:: Interpreter [Google] #HL

  A program that can analyse and execute a program line by line.

* What process specifically does "analyse" mean, why is it needed?
* What does "execute" mean, is :code:`system("/bin/program")` ok?
* "line by line"
* A program that executes a program seems relevant.
* Is "execute" just a synonym of "interpret"?

.. image:: ../deeperMeme.jpg

What is an interpreter ? (II)
-----------------------------

* Which parts are correct, and which are "written by non-programmers"?

.. topic:: Interpreter [Cambridge]

  A computer program that changes the instructions in another program into a form that can be easily understood by a computer.

* "easily understood"?

.. topic:: Interpreter (common definitions)

  A computer program that executes each of a set of high-level instructions before going to the next instruction.

* "high-level instructions", "next instruction" ?

What is a program ? #Cols
-------------------------

* The execution of a program on a machine creates observable effects.
* We can think of it as a mapping from input to output.
* Some output may be explicit, e.g. writing values to a terminal.
* Some output may be implicit, e.g. delays in time.
* The definition of the program is relative to a machine.
* Computers (physical hardware) are *only* one type of machine.

.. image:: WhatIsProgram.jpg

Machine #Cols
-------------

* Mathematical formalism.
* Input is a set of *discrete* values.
* Output is a set of *discrete* observable effects.
* Steps are *discrete* transitions: simple enough to perform mechanically (no "magic").
* "Physically realizable": we could build it.
* Constained by Information Theory / Physics 

.. image:: WhatIsMachine.jpg

What is a language?
-------------------

* Formally: a language is a set of strings.
* Several (related) meanings:

1. The language accepted by a machine: the set of valid inputs.
2. Programming Language: the set of programs that are valid on a machine.

* We'll see the first meaning later in the course: parsing.
* You are expected to have an informal / intuitive understanding of the second.
* We start the lecture series exploring the second meaning.
* What goes into the design of a language / what is the underlying machine.


What is an interpreter ? (III)
------------------------------

.. topic:: Interpreter [me]

  Implementation of a machine that executes programs in a language.

* Every interpreter is a simulation: of a particular machine.
* By simulating the steps of the machine...
* ...updating the state of the simulation...
* ...as defined by the transitions...
* ...we are executing a particular program.
* Consuming input.
* Making (mechanical) decisions.
* Producing output.

Difference between compilers and interpreters #Cols
---------------------------------------------------

* Trivially, speed...
* The number of simulations that we run.
* Interpretation is a simulator running a simulator.
* Compilation is the red dashed relation.
* Convert from a program of one machine (language) into an equivalent program of another.
* Translation is a one-time cost: do not pay overhead of extra simulation.

.. image:: CompilerVsInt.jpg

Difference between compilers and interpreters II #Cols
-------------------------------------------------------

* Compilation is global.
* Interpretation is local.
* The meaning (*semantics*) are fixed at "binding times".
* "compile-time" and "run-time".
* To change compiled code we need to recompile (the whole program).
* To change interpreted code we update the simulation between steps.
* "line by line" -> local steps in the machine simulation.



.. image:: CompilerVsInt.jpg

Binding time #Cols
------------------

* Physically realisable implies no infinite regress in the steps.
* We know the steps terminate - we can do things between them...
* Run another program (embedding).
* Change the code (modding).
* Incremental changes, without restarting the host application.
* Even after shipping - end-user modifications / scripting.

.. image:: LocalGlobalTranslation.jpg


Industrial context
------------------

.. image:: IndustrialContext.jpg

Runtime context
----------------

.. image:: RuntimeProblem.jpg

Synthetic problem
-----------------

.. image:: TeachingProblem.jpg

Course structure
----------------

* Three parts: Lua, Parsing, Integration.

+-----------------------+----------------------------------+
| Part I : Lua          | 4 Lectures                       |         
+                       +----------------------------------+
|                       | Read: Lua                        |         
+                       +----------------------------------+
|                       | Lab: Lua + Irrlicht              |         
+-----------------------+----------------------------------+
| Part II: Parsing      | 4 Lectures                       |         
+                       +----------------------------------+
|                       | Read: Lua                        |         
+                       +----------------------------------+
|                       | Lab: Parser                      |         
+-----------------------+----------------------------------+
| Part III: Integration | 4 Lectures                       |         
+                       +----------------------------------+
|                       | Read: Spark, Piko, Adaptive Rate |       
+                       +----------------------------------+
|                       | *Project*                        |         
+-----------------------+----------------------------------+

Background reading
------------------

* Read it twice: before and after. Why?
* The Lua paper.
* Spark
* Piko
* ACM TOG on adaptive rate.

* What are you learning from the course reading?
* How things should be done.
* How things get done.
* Understanding why the gap.

Programming languages
---------------------

* Why are there so `many <https://en.wikipedia.org/wiki/List_of_programming_languages>`_ different programming languages?
* Inertia
* Not Invented Here
* Commercial Control
* Experiments
* Domain Specialisation

Inertia
-------

* Take a long time to develop a language
* Lua - 1993
* Python - 1989
* Java - 1995
* Julia - 2009
* Rust - 2010
* It takes longer to train programmers, develop common idioms (patterns), build tools and infrastructure, community...

NIH / commercial control
-------------------------

* Consider Java vs C#
* Consider Perl: the panic if the developers realised it was lacking a module...
* Consider Python: the "pythonic" approach.
* Community standards / style guides.
* Objective C / Swift.
* Programmers like writing code.
* Maintaining / merging - not so much.
* Leaving the ecosystem / platform is a (mental) context switch.
* Avoiding switches give a motive to expand ecosystems.
* Incentive to write new code, try new think. Emergent behaviour?

Features
--------

* Sometimes a feature is interesting enough to design a language around it.
* Large-scale experiment: how does it affect software engineering.
* Java - byte codes for platform independence.
* Rust - explicit memory ownership.
* Prolog - automatic unification

* Javascript - interactive web components.
* Make - software dependencies
* C - portable assembly / system programming.

Language design space
---------------------

* Before we look at Lua design - need context.
* Design-space approach: identify relevant language features.
* Treat as dimensions - projecting languages as points into space.
* Relative positions tells us about relationships.

+------------------------------------+------------------------------------------------+
+ Syntax                             + Simplicity vs Power                            +
+------------------------------------+------------------------------------------------+
+ Implementation Cost                + Cheap ... Expensive                            +
+------------------------------------+------------------------------------------------+
+ Runtime Performance                + Low ... High                                   +
+------------------------------------+------------------------------------------------+
+ Data Abstractions                  + Plain (machine-like) vs Rich (domain-like)     +
+------------------------------------+------------------------------------------------+
+ Extensibility                      + Easy / Complete ... Hard / Partial             +
+------------------------------------+------------------------------------------------+
+ Reflection                         + Easy / Complete ... Hard / Partial             +
+------------------------------------+------------------------------------------------+
+ Safety                             + Weak ... Strong                                +
+------------------------------------+------------------------------------------------+

Syntactic complexity
--------------------

.. image:: SyntacticComplexity.jpg

* Measure the number of different syntactic forms (*constructs*).
* Measure minimal size of a specific implementation, e.g. a queue.
* Simplicity is a small number of constructs (robust, elegant).
* Power is the ability to express a wide range succinctly (natural).
* Not really a tradeoff - achieving both is desirable.
* Lisp is minimal: :code:`(map (compose concat tostring) (list 1 " + " 1))`
* Python is expressive: :code:`"".join([ str(x) for x in (1," + ",1) ])`


Implementation cost
-------------------

.. image:: ImplementationCost.jpg

* A self-interpreter in Lisp or Prolog is small.
* Another measure of cost: implementation of `Lisp in Python <http://norvig.com/lispy.html>`_.
* Smaller programs are simpler.
* Is the implementation correct?
* Can the implementation be maintained? Updated?
* Time to port to new architectures, develop new features.

Runtime performance
-------------------

.. image:: RuntimePerformance.jpg

* May be different approaches for same language.
* What kind of optimisations does the language enable?
* How steep is the curve of diminishing returns?
* How old is the language (where are we on the curve) ?
* Vary widely by test case (program) (`Benchmarks Game <http://benchmarksgame.alioth.debian.org/>`_)

Data abstraction
----------------

.. image:: DataAbstraction.jpg

* Low-level abstractions are close to the machine.
* Floats, ints, machine-words: exact binary layout in memory.
* Pointers allow data-structures: tied to exact instantiation.
* High-level abstractions are close to problem domains.
* Strings, dictionaries, references.
* Relationship to efficiency?
* Relationship to productivity?

Extensibility
-------------

.. image:: Extensibility.jpg

* We can alway put more functionality in by adding libraries.
* How difficult is it to extend the syntax / semantics?
* Useful to specialise a language to a domain.
* Custom syntax, structures for particular problems.
* Not just solve problems - clean, simple solutions.
* Who is the programmer - supplier or the user?


Reflection
----------

.. image:: Reflection.jpg

* Extensibility was writing meta-structure into programs.
* Reflection is related: reading meta-structure from programs.
* Examining the data-model in the program at runtime.
* Basing program decisions on properties of the model.
* Inspection, profiling: walk through live data-structures?
* Examine a representation of the code?

Safety
------

.. image:: Safety.jpg

* Can the program break?
* What level of verification can we do?
* Where do errors occur: compile-time, run-time.
* Can we trap run-time errors - do they abort the program?

Visualisation #Cols
-------------------

* Each language is a seven-dimensional point
* How can we view / compare them?
* Radar charts
* Roughly: desirable end of scale is the outside.
* SYN is a special case - both ends desirable in different domains.

.. image:: RadarMap.jpg

C #Cols
--------

.. image:: CLangDesign.jpg

* Domain: system programming.
* Runtime performance is critical.
* Thin layer to the machine (portable assembly).
* Lacks symbolic features (rich abstraction, reflection, extensibility).
* Weak type system, no memory safety.
* No exception system, no modules.
* Procedures for structure, non-composible.


C++ #Cols
---------

* C + classes: still thin machine abstraction.
* Adds templating (partial access to compile-time abstractions).
* Syntactic complexity?
* Still no memory (or strong type) safety.
* Adds RTTI: allows some reflection (access to implicit type tags).
* Are any C++ compilers correct?
* Compilers maturing: performance approaching C.


.. image:: CppLangDesign.jpg


Lisp #Cols
----------

* Radically different point in the space to C/C++.
* Functional programming.
* Symbolic evaluation.
* Homeoiconic : source is its own parse-tree.
* No difference between data and code.
* Programs can be generated dynamically.
* No safety (typing, exceptions, weak bias).
* Add C-style syntax... Javascript.

.. image:: LispLangDesign.jpg

Prolog #Cols
------------

.. image:: PrologLangDesign.jpg

* Comparison to Lisp - very similar.
* Completely different language paradigms.
* Logic programming. 
* Control flow made of search-trees over equations.
* Relatively weak / efficient logic.
* Replace with stronger / slower: theorem provers, sat-solvers...

Haskell #Cols
-------------

.. image:: HaskellLangDesign.jpg

* Functional programming (same paradigm as lisp).
* Statically typed, emphasis on safety.
* Many (strong) compile-time guarantees.
* Requires a clever compiler (ghc is quite mature).
* Only lose 2-3x at runtime.
* Bit more "difficult" to work in (productivity trade-off is complex).

Python #Cols
------------

.. image:: PythonLangDesign.jpg

* Procedural / Functional / OO mixture.
* Design focus on balancing trade-offs.
* Very easy to write code (expressive).
* Slow. (2-100x depending on runtime and domain).
* Dynamic typing.
* Very difficult to maintain code.
* Duck-typing vs transparency / robustness.

Lua #Cols
----------

* Created in 1993 as an embedded scripting language.
* Extend functionality at run-time by loadable scripts.
* Originally for industrial-control.
* Now popular in games industry.
* Two visible interfaces.

1. As a language to write scripts in.
2. As an API to call from an application.

* Simplicity - robustness.
* Uniformity - interchangable.


.. image:: LuaDesign.jpg

What do we need to know?
------------------------

* How to work in Lua - quick language tutorial.

1. Data model.
2. Control-flow.
3. Organisation.
4. Design principles. 

* We are teaching you explicitly how to use Lua.
* Only one specific technology.
* Enough design to generalise to other languages / interpreters.

* How to use the API.

1. Passing control
2. Exchanging data.

State
-----

* What is the state of the system?

.. quote:: 

  The particular condition that something is in at a specific time.

* What would we need to save to resume an operation later?

Low-level view (concrete).

* OS: context-switch between processes.
* CPU-state (PC, registers, stack, page-tables etc).

High-level view (abstract).

* Interpreter holds a representation of the program state.
* Explicitly manipulates it to perform steps in the program.

State II
---------

* Interpreter is simulating a machine.
* What is the state of the machine?

1. Current Location - what do we do next?
2. The values of all variables.

* Current Location can be complex to represent.
* Current statement in program?
* Procedures? We need a call stack to handle returns.
* Objects? What is the current method bound to?
* :code:`if x != y.check()  && !z || flag` ?
* What about conditions, expressions being evaluated?

* We will explain values (data) first, return to control (code) later.

What data is in the program state?
----------------------------------

* Any data that an interpreter would *need* to run the program.

Typical procedural language needs:

* Everything in the local scope.
* Any global data.
* Any calling scopes that may be resumed (returned to).

We can contrast this to C, also needs:

* Any memory that can be reached from a live pointer.
* ...and all the rest of the memory! ( *pointer arithmetic* )
* Projection of bits in memory onto any datatype ( *explicit casting* )
* This is a bit ugly, but possible: `Ch Interpreter <http://www.drdobbs.com/cpp/ch-a-cc-interpreter-for-script-computing/184402054>`_ .

C is not a simple language
--------------------------

* C is a system programming language.
* High-performance, low-level code.
* Full access to the machine.

* Within its domain: explicit memory control is an advantage.
* For high-level scripting: explicit memory control is a disadvantage.
* Recall: all languages are simulations of a machine. 
* Not simulating the memory is the easier choice.
* Scripting domain: telling the host application what to do is important.
* The exact memory contents inside the host process - not so much.
* Needs a more abstract approach.
* Where is the memory explicit in the C language design?

Data model : types
------------------

* The data-model defines all values the programmer may manipulate.
* In imperative languages: state of variables.
* Defines which set of values can be stored in variables.

.. topic:: Type

  A set of possible values, and a definition of the operations that can be evaluated upon them.

* Typical choices in a language depend on the level of abstraction.

.. topic:: Integer (concrete type)

  Bounded set, typically \\( \{ 0 \\leq x \\leq 2^n \} \\) where \\(n\\) is the register size. Operators map onto assembly instructions.

Explicit bit-representations in C #Cols
---------------------------------------

* All data has an explicit representation in bits.
* Every value in every type in C.
* Atomics: fixed number of bytes, specific meaning in each bit.
* All data has an address, direct access :code:`&` and :code:`*`.
* Aggegrate structure built from addresses.
* Casting always possible - access to explit byte sequences.
* *concrete*: machine (platform) details leak into the language.

.. image:: ExplicitCRepr.jpg

What is the alternative to explicit representations? #Cols
----------------------------------------------------------

* Symbolic languages do not tie values to specific bit representations.
* Programmer cannot cast - cannot break the encapsulation.
* Programmer cannot make arbitrary values from raw bits.
* Only constructors can build values.
* Separates the semantic domain of the values from their concrete repr.
* Types are "strong".

.. image:: SymbRepr.jpg

Data model : types II
---------------------

.. topic:: Integer (symbolic type)
 
  Infinite set, \\( \{ \\mathcal\{N\} \} \\), standard arithmetic operators map onto library routines manipulating vector representation of digits.

* Model exposed to programmer is a choice.
* In principle, a type could use any data-structure and algorithms
* Closer to domain: easier to work with (e.g. image, file, sound).
* Closer to machine: easier to execute / faster (e.g. array, int).
* What is the right combination of types to put in a language?


Data model : atomic values
--------------------------

* A value is atomic if it does not contain other values.
* e.g. in C, the :code:`char`.
* Lua provides :code:`number` and :code:`string` as atomics.
* Numbers are double-precision floats (or long integers), no distinction.
* Strings are byte sequences.
* Not null-terminated: :code:`"\0"` is a valid string.
* Explicit conversions, :code:`print(tostring(5))` :code:`print(tonumber("7"))`.
* Not a cast: :code:`print(tonumber("12d"))` produces :code:`nil`.
* Free *coercion* between them: :code:`print("5"+7)` :code:`print(string.reverse(123)`.

Data model : aggregate values
-----------------------------

* Normally language designers supply a range of aggregates.
* Programmers consider datatypes / access-patterns.
* e.g. arrays for data over dense ranges, vectors? lists?
* e.g. dictionaries (maps) for data with sparse key-sets.
* Lua only provides a single type for aggregation: :code:`table`.
* Different syntaxes for construction.

1. :code:`dict = { eggs = 'ham', newblack = 'orange' }`
2. :code:`fib = { 1, 1, 2, 3, 5, 8 }`
3. :code:`mix = { ['eggs'] = 'ham', [5] = 7, 'some', 'more', 2}`

* :code:`print(dict['eggs']) print(dict.newblack) print(fib[2])`
* What gets printed?

Data model : aggregate values II
--------------------------------
 
* Tables cover all use-cases if we ignore efficiency.
* Associative dictionaries are the most expressive type. 

:code:`for k,v in pairs(fib) do print(k,v) end`

.. code::

  1	1
  2	1
  3	2
  4	3
  5	5
  6	8

* Kind of weird for a programmer, normal for a mathematician.
* We can override it :code:`x={ [0]=1, 1, 2, 3}`. 
* The issue is somewhat `controversial <http://lua-users.org/wiki/CountingFromOne>`_. 

Data model : aggregate values III
---------------------------------

.. code::

  mix = { ['eggs'] = 'ham', [5] = 7, 'some', 'more', 2}

.. code::

  1	some
  2	more
  3	2
  eggs	ham
  5	7

* Dense key-values (arrays) are just a special case.
* Size is accessed by the :code:`#` operator.
* Appending to a dense array: :code:`x[#x+1] = y`.
* No error if keys do not exist: :code:`print(mix.blah)` produces :code:`nil`.
* Because all values are first-class, tables can also be keys...
* Could not think of a use for this.. but hey, it's nice!

Static vs dynamic types #Cols
-----------------------------

* Weak types: any memory can project into any type.
* Strong types - we must make a choice, do we store tag?

1. Associate the type with variable (static)
2. Associate the type with value (dynamic)

* Difference: :code:`x=7 x="hello"`
* In dynamic case, variables are only names.
* Scopes are then just tables...

.. image:: TypingOptions.jpg

Type system in lua
------------------

* All values are **first-class** : no special per-case rules.
* Design principle: uniformity / regularity is simpler.
* Store in a variable, pass as an argument, return as result.

+-------------------------+-------------------------+
| Atomic: nil             + Empty                   |
+-------------------------+-------------------------+
| Atomic: boolean         + true false              |
+-------------------------+-------------------------+
| Atomic: number          + double                  |
+-------------------------+-------------------------+
| Atomic: string          + byte seqeunces          |
+-------------------------+-------------------------+
| Atomic: function        + args + code + ret       |
+-------------------------+-------------------------+
| Atomic: userdata        + byte arrays (opaque)    |
+-------------------------+-------------------------+
| Atomic: thread          + active control          |
+-------------------------+-------------------------+
| Aggregate: table        + pairs                   |
+-------------------------+-------------------------+

Data model : tables everywhere #Cols
------------------------------------

* Tables are the only aggregate, look like records.
* So where we see :code:`io.write(x)` we can ask :code:`print(type(io))`.
* So what is in the module (table)?  

:code:`for k,v in pairs(io) 
do print(k,v) end`.

* Explore the system interactively.
* :code:`file` was not a basic type...

* If try to list :code:`pairs(io.stdin)` ...

.. code::

  lines	function: 0x1045143e2
  type	function: 0x1045146e1
  stderr	file (0x7fff711133e0)
  stdin	file (0x7fff711132b0)
  stdout	file (0x7fff71113348)
  read	function: 0x10451464f
  popen	function: 0x1045145aa
  write	function: 0x10451474c
  close	function: 0x104514330
  open	function: 0x10451448b
  ...

Data model : examples
-------------------------------------

* We get an error message:

.. code::

  stdin:1: bad argument #1 to 'pairs' (table expected, got userdata)

* So, opaque data - we've hit the C ABI interface.
* Data-structures? Lots shown in `PIL <https://www.lua.org/pil/contents.html>`_ (chapter 11).
* Basic idea: everything is a table, pick record or array syntax as appropriate.
* Example: matrices, :code:`x = { {1,2}, {3,4}}`.
* Access works as expeced, e.g. :code:`print(x[1][1])`.
* Lists? :code:`element = { data='blah', next=nil } element.next={data=2,next=nil}`.
* Trees? :code:`node = { data=3, children={} }`.

Data model : arithmetic
-----------------------

* All numbers are floating point (double-precision 64-bit).
* Binary operators: :code:`+`, :code:`-`, :code:`*`, :code:`/`, :code:`^`.
* Normal floating point issues

.. code::

  > print(1/3*10000 - 3333)
  0.33333333333303

* Rounding issues do not apply to integers or `dyadic rationals <https://en.wikipedia.org/wiki/Dyadic_rational>`_.
* Modulus is generalised to floats: :code:`2.75 % 0.5 == 0.25`.
* Modulus handles negative values correctly (unlike C), e.g. :code:`-3 % 2 ==1`.

Data model : comparisons
------------------------

* Simple types (numbers and strings) compare values.
* Equality is exact - standard floating point issues apply.
* Inequality is :code:`~=`.
* Ordering is lexigraphic for strings, standard for numbers, :code:`<`, :code:`>`, :code:`<=` and :code:`>=`.
* Avoiding ambiguity - no coercion from strings to numbers for comparison.
* Corner-case is :code:`2<15` (standard numerical ordering), but :code:`"2">"15"` (lexigraphic).
* Tables, userdata and functions are equal by reference (same object).
* :code:`a={1} b={1} print(a==b)` produces?

String processing
-----------------

* Declaration: three quote types to avoid escaping.
* Concatentation: :code:`..` operator.
* Avoids ambigiuity between :code:`print("3"+"4")` and :code:`print('3'..'4')`.
* Substrings: :code:`print(string.sub("abcdef",2,-2)` (count from 1).
* Repetition: :code:`x = string.rep('\0',2^20)` (zero'd mb of memory).
* printf formating: :code:`s = string.format('%s,%02d','hello',12)`.
* Decoding strings: :code:`print(string.byte("a"))`.
* Encoding strings: :code:`print("easy as "..string.char(97,98,99))`.
* Simple search: :code:`print(string.find("a simple string","imp"))`.
* Replace: :code:`print(string.sub("abc","b","Z"))`.
* Patterns: :code:`print(string.find("123+4/3","%d+%D"))`.



Code model: simple statements
------------------------------

* Control-flow in Lua is a simple procedural language.
* **Chunks** are sequences of Lua statements.
* Semicolons are optional: syntax of each statement is *self-delimiting*.
* Assignment works in parallel: both targets are sources are sets.

:code:`x, y[1] = "hello", x+7` both occur at same time.

* The evaluation of the expressions on the r.h.s. occurs first.
* :code:`x+7` is evaluated before :code:`x` is written into.
* Swaps work: :code:`x,y = y,x`.

* Global variables are created on assignment.
* Non-existent variables evaluate to :code:`nil`, to delete :code:`x = nil`.

Code model: scopes
-------------------

* Each chunk in the code has its own scope.
* Each scope is a table, default target is the global scope :code:`_G`.
* The local keyword writes into the chunk's own scope, :code:`local x=3`.
* To query the global scope, just a table: :code:`for k,v in _G do print(k,v) end`.
* Querying the local scope is a little more involved.
* The local scopes are stored in the call-stack, can access *directly*.
* Very different to C, where the call-stack is undefined / platform-specific.
* :code:`debug.getlocal(1,n)` will get local name,value at index n.
* The index 1 means the top of the stack (we can also access caller's scopes...)

Code model: functions
----------------------

* Function calls are similar to C syntax: name parentheses arguments.
* Weirdly the parentheses are *optional* if there is a single argument.
* As is whitespace because the syntax is self-delimiting.

.. code::

  print("hello ",3,' and ',21)
  print"yo"
  print {1,2,3}
  print (1,2)

* Definitions can look normal:

.. code::

  function makeLabel(name) return string.format("lab=%s",name) end

Code model: functions II
-------------------------

* But functions are really just values, so definitions can look weirder:

.. code::

  makeLabel = function (name) return string.format("lab=%s",name) end

* When we define a function this way it is anonymous (lambda-expression).
* The resulting function value is then being named by assignment.
* Access to lambdas means that we can build arbitrary despatch logic.
* e.g. list of processing calls: :code:`doWork = { function(x) ... end, function(x) ... end }`.

Code model: functions III
--------------------------

* e.g. tables of functionality: :code:`blah = { cons=function(x,y) ... end, update=function ... }`.
* These start to look like libraries / packages...
* The built-in libraries are just tables of functions :code:`print(type(io))`.
* Language uniformity produces simplicity - fewest, most powerful mechanisms.
* So what about OO? It's just another form of packaging...

.. code::

  Account = {balance = 0}
  function Account.withdraw (v)
    Account.balance = Account.balance - v
  end

* No special case on function name, writing into table directly.

Code model: functions IV
-------------------------

* Example only worked on the "object" :code:`Account`, name was hardcoded.
* If we use a table as an object we need to specify this / self to method.

.. code::

  function withdraw (self,v)
    self.balance = self.balance - v
  end
  Account = {balance = 0, withdraw=withdraw}
  Account.withdraw(Account,10)
  Account:withdraw(10)          -- equivalent form

* The colon is **syntactic sugar** - hides the first argument.
* Can use it in calls / definition (produces the name "self").

Code model: other OO functionality
-----------------------------------

* The idea behind OO is code-reuse, normally via inheritence.
* If a subtype lacks specific functionality, reuse the supertype.

.. topic:: Metamethod

  Each of the standard operators in Lua can be overriden (despatched to a custom method) called a `metamethod <http://lua-users.org/wiki/MetamethodsTutorial>`_ .

* e.g. if we wanted to customise addition :code:`a+b`, :code:`setmetatable(a,{__add=f})`.
* Looking up a name in a table is an operator called index.
* Specific detail: :code:`__index` is called if the name is not found (i.e. check table first, then call metamethod if name is missing).
* This allows us to build up an inheritence hierarchy dynamically.
* Contrast to duck-typing in Python, prototypes in Javascript.

Code model: OO example
-----------------------

* To define a "class" we can use...

.. code::

  function Classname:new()
    res = {}
    setmetatable(res,self)
    self.__index = self
    return res

* We could also avoid smashing together the metatable and class namespace.
* Only useful if we wanted to define operations on classes separate from instances.
* Lua is not an OO language - primitives are powerful enough to build our own OO.

Code model: imperative structures
----------------------------------

* Skipped basic control-flow until now (functional style is universal).
* Simulate conditional expressions: :code:`x and "truecase" or "falsecase"`.
* But imperative styles are terse, useful in scripts.
* :code:`if expression then chunk else chunk end`.
* :code:`while expression do chunk end`.
* :code:`repeat chunk until expression`.
* C-style for-loops :code:`for i=1,n,step do chunk end`.
* for-each iterators: :code:`for x in expression do chunk end`.
* Recall: each chunk is both a sequence of statements *and* a local scope.

Code model: nil in expressions
-------------------------------

* An expression is not always an explicit comparison, e.g. :code:`if x<3 do ... end`.
* When the comparison is not explicit there is an implicit comparison to nil.
* e.g. :code:`if x then ... end` means :code:`if x~=nil then ... end`.
* In general we think of :code:`nil` as "undefined" or "doesn't exist".
* Returning to the example of querying the local scope.
* :code:`debug.getlocal` used indices, no explicit check on length.

.. code::

  i=1
  while debug.getlocal(1,i) do ... end

Code model: coroutines
-----------------------

* Coroutines are a model of concurrency, somewhat similar to threads.
* Threads are multiple flows of control with a shared memory.
* Coroutines are multiple flows of control *without* a shared memory.
* Instead routines (procedures) cooperate via a call/return mechanism.
* Each coroutine can *yield* control: returning a value.
* When called again they resume from the point that they yielded.
* Context has been saved, and is reloaded on the next call.

Code model: standard call #Cols
--------------------------------

* Normal call sequence creates a scope (every chunk makes a scope).
* Single stack of calls.
* Push scope on a call.
* Pop (destroy) scope on return.
* For coroutines we need the scope to live on somewhere.
* Multiple stacks.

.. image:: NormalFunction.jpg

Code model: simple coroutine example #Cols
--------------------------------------------

.. code::

  function foo()
    print("foo", 1)
    coroutine.yield()
    print("foo", 2)
   end

  co = coroutine.create(foo)
  coroutine.resume(co)
  > foo      1
  coroutine.resume(co)
  > foo      2

* The coroutine package lets us build threads.
* Each thread has its own stack.
* Inside the thread, yield suspends the thread.
* Resumes the calling thread.
* Calling resume again picks up at the same point.

Code model: building generators #Cols
-------------------------------------

* Build it: :code:`c=coroutine.create(counter)`
* Start it: :code:`print(coroutine.resume(c,9)`
* yield args = resume results.
* resume() args = yield results.

.. code::

  function counter(arg)
    local i=arg
    while true do 
      coroutine.yield(i)
      i=i+1
    end end #NoHL

.. image:: Coroutine.jpg

Code model: wrapping as iterators
---------------------------------

* Iterator function return a wrapper for calling a coroutine.
* Can be used in :code:`for x in f() do ... end`.
* Wrapper function calls coroutine until exhausted, returns nil.

.. code::

  function f()
    -- build a coroutine from fbody called c
    return function() return coroutine.resume(c) end

* This is common enough to be supplied, :code:`coroutine.wrap(fbody)`
* Quite a neat permutations example in `PIL9.3 <https://www.lua.org/pil/9.3.html>`_.
* Quick mention: :code:`function(...)` defines a `variadic function <https://www.lua.org/pil/5.2.html>`_.
* Collects arguments into table called arg.



Language Extension
------------------

* So we have designed a nice shiney new language.
* Lots of lovely symbolic types, fits well to a problem domain.
* Then an awkward user appears, wants to do X.
* There is no facility for accessing X in the language.
* Is the language **extensible**, can the user extend it themselves?
* We can do anything from C - low-level / universal.
* If we can glue bits of C code into programs, universal interface.
* Typical approach is a foreign function inteface (FFI).

Foreign Function Interface
--------------------------

Python as an example:

.. code:: 

  import ctypes
  ext = ctypes.CDLL( 'libuser.so' ) # load dynamic link library
  x = ext.func("hello")             # change of language

* The FFI is a bridge to code written in another language.
* Somewhere there is :code:`int func(char *)`.
* Both languages share concept of procedures / calls.
* Pass control from one to another.
* Values can cross the bridge.
* Need to be converted: :code:`str` object <-> :code:`char *`.

Extension via a FFI #Cols
-----------------------------

* Extending a language allows calls out.
* Host: Lua,   Foreign: C.
* Implies that :code:`main()` lives in the host.
* The application needs to be written in Lua.
* Specific parts can be pushed out to the foreign language.

.. image:: LanguageExtension.jpg

Extension vs Embedding #Cols.
-----------------------------

.. image:: LanguageEmbedding.jpg

* The other way around.
* Now the host language is C, bulk of application.
* Control originates in :code:`main()` in C.
* Passes control to Lua to perform specific tasks.
* No need to define an FFI within Lua.
* Instead, define an API to call Lua functionality from C.
* Lua is now **embedded** inside the C language.

Embedding Lua
-------------

* Only providing eval: pass code as a string.
* Code is executed by the interpreter, updates the Lua state.
* Allows assignment (build values), query (return values), calls.

.. code::

  #include <stdio.h>
  #include <lualib.h>
  int main (void) {
      lua_State *L = luaL_newstate();   // empty state
      lualL_dostring(L, "x=3 y='hello'")
      lua_close(L);
      return 0; 
  } 

Minimal Embedding
-----------------

* Efficiency - manipulating strings in host to pass to embedded.
* Cumbersome - string processing for all ops.
* Serialisation (and parsing) needed at interface point...
* Wrong kind of simplicity - the interface is minimal.
* Everything that touches the interface becomes complex.

.. code::

  char luacmd[128];
  sprintf(luacmd, "x = %d", x);
  luaL_dostring(L, luacmd);

Typed Interface #Cols
---------------------

* We do not want to communicate using strings: avoid serialisation.
* How to convert C datatypes into Lua datatypes?
* How to pass converted values to the interpreter?
* How to receive Lua values from the interpreter?
* How to convert Lua values into C datatypes?
* Need an API to call using those types to embed Lua in C.

.. image:: DataFlow.jpg

Simple API
----------

* It is essential that we keep the API as small as possible.
* Large APIs are more difficult to use / maintain.
* The first part that we look at is calling Lua function from C.
* To be useful the function needs arguments to pass data in.
* A simple approach is to fix a data mapping, e.g.

+------------------+--------------------+
| Lua datatype     | C datatype         |
+------------------+--------------------+
| number           | :code:`double`     |
+------------------+--------------------+
| string           | :code:`char *`     |
+------------------+--------------------+

* Already we run into a problem: which side manages the string memory?
* Return to that issue later - now we look at calling functions...

Simple API 2
------------

* If we want to call function in Lua we need a procedure.
* C procedures are typed by their arguments.
* So we need something like:

.. code::

  void call_func1n(double a1);
  void call_func1s(char *a1);
  void call_func2nn(double a1, double a2);
  void call_func2ns(double a1, char *a1);
  ...

* Already it gets quite ugly. 
* For \\(n\\) datatypes and up to \\(k\\) arguments: \\(\\mathcal{O}(n^k)\\) variations.

Simple API 3 #Cols
------------------

* We could try to hide this complexity in the data-structure.
* Add a bit more wrapping around the data.
* The :code:`union` can be read as "one of".
* The value in the tag tells us which one is valid.
* The programmer has to track these assumptions (bugs).

.. code::

  typedef struct _Data {
    int tag;
    union {
      double num;
      char *str;
    } u;
  } Data;

  Data a1 = { 0, 0.123 };
  call_func1(&a1);

Simple API 4
------------

* We've lost simple declarations in expressions.
* Explicit declaration and dereference (clunky and verbose).
* The API is now down to \\(\\mathcal{O}(k)\\) variations.
* Memory issue: who owns it? when is it safe to deallocate?
* Interaction with GC.

Stack Approach #Rows
--------------------

* Interface between the host and the interpreters is a stack.
* Interpreter will use it strictly as a stack (push/pop top).
* Host has more flexibility to rearrange items.
* Each value is a lua object.
* Lifetime of the values is while they are on the stack.

.. image:: CallViaStack.jpg


Stack Approach
------------------

* Lua objects are opaque - we do not need to see the struct definition.
* The stack becomes the entire interface.
* No need to map runtime assumptions (enums) onto data.
* More robust approach.
* Need one API call per datatype (push).
* Need API calls to locate functions (first class values).
* API call to execute a function call from the stack contents.

.. code::

  lua_getglobal(L, "f");
  lua_pushnumber(L, 1.234);
  lua_pushstring(L, "hello world");
  lua_pcall(L, 2,0,0);

Lifetimes
---------

* The arguments passed are on the stack during pcall execution.
* The C code is offering a guarantee to the Lua interpreter.
* All memory will remain valid during the length of the call.
* Return values from Lua functions are added to the stack.
* They remain live until they are removed.
* After they are popped from stack - Lua can garbage collect them.

.. code::

  lua_pcall(L, 2,1,0);
  double res = lua_tointeger(L, -1);
  lua_pop(L, 1);

Stack API
---------

* We can break down the parts of the API that we've seen into groups

+---------------------+------------------------------+
+ lua_pushX           | convert type X from C to Lua |
+---------------------+------------------------------+
+ lua_toX             | convert type X from Lua to C |
+---------------------+------------------------------+
+ lua_pop             | delete top of the stack      |
+---------------------+------------------------------+
+ lua_pcall           | call Lua, specified on stack |
+---------------------+------------------------------+

* The lua_toX functions do not change the stack, only *peek*.
* Decoupling from pop allow inspection without destruction.
* Otherwise the pop step would need to handle memory management.

Building Complex Values
-----------------------

* Aggregate values are built up as a series of stack operations.
* Example: :code:`x = { n=2, name="alpha", cnt={11,22} }`
* We want to end up with the table value on the stack.
* :code:`lua_settable(L,n)` writes a k=v pair into the table at n.
* To build each pair we push k, then v.
* If we need building steps then after executing them we need to end with k,v at the top of the stack.

.. code::

  lua_newtable(L);         // Outer table
  lua_pushstring(L,"cnt"); // Need k,v later
  lua_newtable(L);         // Inner table stored in cnt
  ... (continued on next slide) ...

Building Complex Values II
--------------------------


.. code::

  lua_newtable(L);         // Outer table
  lua_pushstring(L,"cnt"); // Need k,v later
  lua_newtable(L);         // Inner table stored in cnt
  lua_pushinteger(L,11);
  lua_rawseti(L,-2,1);     // t[1] = int(11)
  lua_pushinteger(L,22);
  lua_rawseti(L,-2,2);     // t[2] = int(22)
  lua_settable(L,-3);      // outer.cnt = {11,22}
  ... similar code for "n" and "name" keys ...

Calling in the other direction
------------------------------

* We can add custom functions to the Lua interpreter.
* These can then be called by any Lua code that is run.
* A function is a primitive type in Lua, interface is C API stack.
* Private stack for call to C (only arguments on it).
* C procedure pushes return values before exit.

.. code::

    static int l_sin (lua_State *L) {
      double d = luaL_checknumber(L, 1);
      lua_pushnumber(L, sin(d));
      return 1;  /* number of results */
    }
    ...
    lua_pushcfunction(l, l_sin);
    lua_setglobal(l, "mysin");


What we do in the first lab
---------------------------

* Low-level picture of execution inside the interpreter.
* How it interacts with control-flow in the host app.
* Event based despatch.
* Select, polling.
* Blocking.
* The Lua Source / Packaging / Linking.
* Assessment will be during the session: i.e. you must attend.

