imgur-gallery-rss
=================

Golang setup a simple server to generate the rss feed for the imgur gallery

My Imgur Gallery RSS feed all of a sudden stopped.

I decided to write my own that parsed the Imgur API.

Runs on port 8080 (yes of course you can change it).

Requires that you register an app with Imgur and export the client ID as an environment variable.

```
export IMGUR_CLIENT_ID=<some_id_of_yours>
```

Then to run it:

```
> go install github.com/quekshuy/imgur-gallery-rss
> imgur-gallery-rss

```
