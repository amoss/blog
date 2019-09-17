# blog
This is the code that drives [my blog](https://mechani.se). It is designed to run inside a jail 
sitting behind an nginx front-end that takes care of TSL. When it is running it expects a 
`data/` directory in the current directory. This should initially contain the files in `src/blog/data/`
and then blog posts as .rst files. Any images used in the post should be placed into a directory
with the same base name as the post, i.e.:

```
blog
data/apost.rst
data/apost/image.jpg
```
