rebuild:
	bundle exec compass compile src
	bundle exec jekyll build

server:
	bundle exec compass watch src &
	bundle exec jekyll serve --watch
