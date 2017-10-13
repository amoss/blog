========================================================
Lab 1: The Lua Interpreter and Irrlicht
========================================================

:Author: Dr Andrew Moss
:Date: 2017 April 12
:CourseCode: DV1564
:CourseName: Interpreters and Scripting
:Style:      Page

Purpose
------------

The purpose of today's session is to gain familiarity with the 
linux environment in the labs, the compiler tool-chain and 
building against source packages.

Assessment
------------

Assessment will be by demonstrating the working code to the 
teacher in the lab.

Introduction
------------

You will build a prototype project that includes both the Lua
interpreter and the Irrlicht graphics engine in the same 
binary. The binary will be statically linked against these 
components, which will be built directly from unaltered source.
You will then write some simple glue code to show that they 
are both working.

Your directory structure should look like this:

.. shell::

  lab1/
  lab1/lua/
  lab1/src/

The latest version of Lua is available from `the download page <https://www.lua.org/download.html>`_. You can fetch in the terminal using:

.. shell::

  curl https://www.lua.org/ftp/lua-5.3.4.tar.gz -o lua.tgz
  tar xzf lua.tgz
  ls lua-5.3.4

Although you may find it more convienient to unpack directly onto the disk:

.. shell::

  curl https://www.lua.org/ftp/lua-5.3.4.tar.gz | tar xz

Taking a look through the lua distribution, a good starting
point is normally the readme file:

.. shell::

  cat lua-5.3.4/README

In this case it is not so helpful so read the fuller version:

.. shell::

  firefox lua-5.3.4/doc/readme.html

The important part is how to build the interpreter vs how to 
install it. We want to produde a library that we can statically link
against - we do not want a target user to need to change their
machine configuration to include Lua so that we can use it. This
is an approach to avoiding `Dependency Hell <https://en.wikipedia.org/wiki/Dependency_hell>`_ that works because the library that we
rely on is very small.

To build Lua we execute :shell:`make linux` inside the lua directory. After the build process completes we want to know what has
changed inside the directory tree. The alternative approach is to
dig through the documentation and try to find a description of how
we link against it - but this way is a bit quicker and guaranteed
to get the correct result. Documentation can often be out of date
or incomplete.

.. shell::

  make clean; find . -type f -exec ls -l \{\} \; | sort -k9 >before.txt; make linux; find . -type f -exec ls -l \{\} \; | sort -k9 >after.txt; diff before.txt after.txt

If you have vimdiff installed then the output is easier to read
than diff, but it is not always a standard package. To quit from
vimdiff type :shell:`:qall!`. Scanning through the output we can
see that the build created many object files, two executables and
a library:

.. code::

  > -rwxrwxr-x 1 amoss amoss 257264 apr  3 09:45 ./src/lua
  > -rwxrwxr-x 1 amoss amoss 175648 apr  3 09:45 ./src/luac
  > -rw-rw-r-- 1 amoss amoss 441924 apr  3 09:45 ./src/liblua.a



Task 1 : Building against the Lua interpreter
---------------------------------------------

The Lua interpreter uses a a :code:`lua_State` structure to 
main the state of the machine. This is passed as a parameter
to most of the Lua C API calls. As an example the following
code will load and execute a Lua script:

.. code::

  #include<lua.hpp>
  #include<lauxlib.h>

  int main(int argc, char** argv)
  {
  lua_State* L = luaL_newstate();
      luaL_openlibs(L); 
      if( argc>1 )
          luaL_dofile(L,argv[1]);
  }

Compilation in the linux environment uses g++, if you save the above code into a file calle "bareloop.c" then it can be compiled with:

.. shell::

  g++ bareloop.c -Ilua-5.3.4/src -Llua-5.3.4/src -llua -obareloop -ldl

To execute the result use an explicit path:

.. shell::
  
  ./bareloop

Look at the API reference for `luaL_dostring <https://www.lua.org/manual/5.3/manual.html>`_ and rewrite
this into an interactive loop that performs a single statement
at a time. Read one line at a time from stdin using normal blocking I/O, and pass it to Lua for execution. Verify that your interpreter performs the same way as the standard interpreter.

Switch your program to use asynchronous I/O instead of blocking
operations. Do not look for an asynchronous read operation, use
the :code:`select()` call to check for input before calling :code:`read()`. A quick guide to using select can be found `here <http://www.tutorialspoint.com/unix_system_calls/_newselect.htm>`_.

When setting the time out, use a value of :code:`1000000/60` to lock
the loop at 60hz, this will make life easier later in the task "60fps, it's not just a good idea".

The final part of the task is to arrange your work so that you 
can use the lua code distribution untouched (without any edits),
and just link against the output library explicitly. You should
wrap this up inside a Makefile (`quick example <http://www.cs.colby.edu/maxwell/courses/tutorials/maketutor/>`_). Be aware of the 
classic "gotcha" for makefiles - the line with the comands after
the rule definition line *MUST* start with a tab. It cannot
use spaces for indentation.

Your makefile should have a rule for the target executable that
depends on your source and the lua library. The rule for building
the lua library should execute make on the lua makefile with the
appropriate target. You will run into `this issue <http://stackoverflow.com/questions/6524771/makefile-confusions>`_ the first time 
you try it.

Optional: because the Lua distribution is untouched there is no
need to keep track of it in your code: either as a final distribuion 
if work with open-source, or in version-control while coding. Work 
out how to put in a rule that will download and unpack the Lua
code if it is not in the appropriate directory on the disk.


Task 2 : Building against Irrlicht
----------------------------------

The second task is to get Irrlicht up and running in the same style:
no modification to the systems, static linking of what is necessary
to pull the library into your code. We have installed the GL libraries
that you will need on the Ubuntu image running in the labs.

The Irrlicht download page  
`is here <http://irrlicht.sourceforge.net/?page_id=10>`_, 
we tested the lab against the 1.8.4 version of the SDK.

.. shell::

  curl -L -oirr.zip http://downloads.sourceforge.net/project/irrlicht/Irrlicht%20SDK/1.8/1.8.4/irrlicht-1.8.4.zip
  unzip irr.zip

You will need to follow the same process as before - work out how to 
build the linux static lib. For the glue code to test it use the first
step of the Irrlicht `tutorials <http://irrlicht.sourceforge.net/docu/example001.html>`_.
You can ignore the crap about WinMain as you will not be building a 
windows version, and rather than smashing together the namespaces it
is slightly cleaner to lookup each of the calls and use explicit namespaces
in the code. The result should be about 50 lines of code.

The build step should look like this, and when you run it you should get
a simple character model spinning in a window.

.. shell::

  g++ task2.cc -Iirrlicht-1.8.4/include/ irrlicht-1.8.4/lib/Linux/libIrrlicht.a -L/usr/lib/x86_64-linux-gnu/mesa -lGL -lX11 -lXxf86vm -otask2



Task 3 : Integrating both libraries with some glue code
-------------------------------------------------------

Integrate both of the previous codes: your main loop should use select()
to check if data is available at the terminal. The timeout should be
set so that the loop will run at 60Hz. When data is available, execute it
in the Lua interpreter. Call the code to render a frame in Irrlicht from
the main loop so that the window with the animated mesh can be seen, while
the interpreter can be interacted with from the terminal.

In `tutorial 4 <http://irrlicht.sourceforge.net/docu/example004.html>`_
it is shown how to update the position of the node with the animated mesh
inside the scene. Update your code to store a :code:`core::vector3df` with
the mesh position.

Using the techniques described in the lectures (and documented in Chapter
4 of the Lua reference manual):

* Implement reflection - add a cfunction that can be called from Lua that
  builds a table of three numbers from the coordinates in the node. Store
  this in the global scope so that you can run the example below.
  
* Implement an update - again a cfunction that can be called from Lua that
  updates the coordinates, e.g. :code:`updatepos(0,0,5)`.

.. code::
 
  for k,v in pairs(getpos()) do print(k,v) end
  1  0
  2  0
  3  0
  updatepos(0,-3,0.5)
  for k,v in pairs(getpos()) do print(k,v) end
  1  0
  2  -3
  3  0.5



