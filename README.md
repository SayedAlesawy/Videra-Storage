# Videra-Storage
![Build](https://github.com/SayedAlesawy/Videra-Storage/workflows/Build/badge.svg?branch=master)

The Storage Module Implementation for the Videra Video Indexer Engine.

## Frontend Contract

### Tags endpoint
```
GET: /tags
```
```
[
  "tag1",
  "tag2",
  "tag3",
  ...
]
```

### Search endpoint
```
GET /search?tag=tag1&start=1&end=6
```
Note: `start` and `end` are optional params
```
[
  {
    "name": "video1",
    "token": "token1",
    "thumbnail": "link/to/thumbnail1",
  },
  {
    "name": "video2",
    "token": "token2",
    "thumbnail": "link/to/thumbnail2",
  },
  ...
]
```

### Stream endpoint
```
GET /stream?token=token1&tag=tag1&start=1&end=6
```
Notes:
- `tag` same as the one sent in `/search`.
- `start` and `end` are optional params (same as the ones sent in `/search`).
- `token` is the video token sent in the `/search` response.
```
{
  "src_link": "link/to/stream/src", //To be passed to the hls.js library
  "clips": [
    {
      "start": 1, //clip start
      "end": 10 //clip end
    },
    {
      "start": 15, //clip start
      "end": 20 //clip end
    },
    ...
  ]
}
```