Getting started with the frontend
---------------------------------

The templates and source posts and pages are in the `src` directory; the Jekyll-generated site is written out to the `site` directory.

Install Jekyll and dependencies. This assumes that you have Ruby, RubyGems, and Bundler installed.

 * `bundle install`
 * `gem install jekyll --no-rdoc --no-ri`
 * `gem install compass --no-rdoc --no-ri`

To rebuild the site:

 * Type `make` and hit return.

To run a local server and have it rebuild the site automatically during
development:

 * Type `make server` and hit return. By default, the dev server is available
   at [http://localhost:4000/](http://localhost:4000/).
