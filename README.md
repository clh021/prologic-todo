# todo
[![Build Status](https://cloud.drone.io/api/badges/prologic/todo/status.svg)](https://cloud.drone.io/prologic/todo)
[![GoDoc](https://godoc.org/git.mills.io/prologic/todo?status.svg)](https://godoc.org/git.mills.io/prologic/todo)
[![Go Report Card](https://goreportcard.com/badge/git.mills.io/prologic/todo)](https://goreportcard.com/report/git.mills.io/prologic/todo)
[![CodeCov](https://codecov.io/gh/prologic/todo/branch/master/graph/badge.svg)](https://codecov.io/gh/prologic/todo)
[![Sourcegraph](https://sourcegraph.com/git.mills.io/prologic/msgbus/-/badge.svg)](https://sourcegraph.com/git.mills.io/prologic/msgbus?badge)
[![Docker Version](https://images.microbadger.com/badges/version/prologic/todo.svg)](https://microbadger.com/images/prologic/todo)
[![Image Info](https://images.microbadger.com/badges/image/prologic/todo.svg)](https://microbadger.com/images/prologic/todo)

todo is a self-hosted todo web app that lets you keep track of your todos in a easy and minimal way. 📝

## Screenshots
_Nord Theme_

<img src="screenshots/mobile-nord.png" alt="Mobile Nord Theme" height="500"/>
<img src="screenshots/desktop-nord.png" alt="Desktop Nord Theme" height="500"/>
<br />

_Dracula Theme_

<img src="screenshots/mobile-dracula.png" alt="Mobile Dracula Theme" height="500"/>
<img src="screenshots/desktop-dracula.png" alt="Desktop Dracula Theme" height="500"/>

See all themes in the "Preset Color Themes" section below

## Demo
There is also a public demo instance avilable at: [https://todo.mills.io](https://todo.mills.io)

## Deployment

### Docker Compose
`docker-compose.yml`
```
version: '3'

services:
  todo:
    image: prologic/todo
    container_name: todo
    restart: always
    ports:
      - 8000:8000
    volumes:
      - todo_db:/usr/local/go/src/todo/todo.db

volumes:
  todo_db:
```
This file:
* Creates the `todo` container using the latest image from `prologic/todo` in [Docker Hub](https://hub.docker.com/r/prologic/todo).
* Binds port 8000 on your host machine to port 8000 in the container (you may change the host port to whatever you wish).
* Volume mounts the database path, saving your todo items so that your todo list will be saved in between container restarts.

Bring the container up with:
```
$ docker-compose up
```

### Docker
Alternatively, you can run the container without docker-compose:
```
$ docker run -p 8000:8000 -v todo_db:/usr/local/go/src/todo/todo.db prologic/todo
```

## Configuration

### Preset Color Themes
todo comes with 12 different color themes based on some of the most popular programming themes:
```
ayu, dracula, gruvbox-dark, gruvbox-light, lucario, monokai, nord, solarized-dark, solarized-light, tomorrow, tomorrow-night, zenburn
```

You can set the theme by passing the `THEME` environment variable to the docker container, for example:

`docker-compose.yml`
```
version: '3'

services:
  todo:
    image: prologic/todo
    container_name: todo
    environment:
      THEME: ayu
    restart: always
    ports:
      - 8000:8000
    volumes:
      - todo_db:/usr/local/go/src/todo/todo.db

volumes:
  todo_db:
```

#### Screenshots

You can find screenshots and the CSS of all the preset color themes in both mobile and desktop views on the [Wiki Page](https://git.mills.io/prologic/todo/wiki/Color-Themes).


### Custom Color Themes
You can set your own color theme by passing in the appropriate environment variables.

Set the `THEME` environment variable to `custom`, and the six following environment variables to the colors of your choice (in hex format, omitting the `#`):

| Environment Variable           | Description                       |
|--------------------------------|-----------------------------------|
| PAGEBACKGROUND                 | Web page background               |
| INPUTBACKGROUND                | Text boxes and buttons background |
| FOREGROUND                     | Input and item text               |
| CHECK                          | Check mark on button              |
| X                              | X mark on button                  |
| LABEL                          | Heading text and button hover     |

An example configuration:

`docker-compose.yml`
```
version: '3'

services:
  todo:
    image: prologic/todo
    container_name: todo
    environment:
      THEME: custom
      PAGEBACKGROUND: 282a36
      INPUTBACKGROUND: 44475a
      FOREGROUND: f8f8f2
      CHECK: 50fa7b
      X: ff5555
      LABEL: ffffff
    restart: always
    ports:
      - 8000:8000
    volumes:
      - todo_db:/usr/local/go/src/todo/todo.db

volumes:
  todo_db:
```

### Additional Configuration
| Environment Variable           | Description                                      | Default Value |
|--------------------------------|--------------------------------------------------|---------------|
| MAXITEMS                       | Maximum number of items allowed in the todo list | 100           |
| MAXTITLELENGTH                 | Maximum length of a todo list item               | 100           |

## Development / Non-Dockerized Deploy
You can quickly run a todo instance from source using the Makefile:
```
$ git clone https://git.mills.io/prologic/todo.git
$ cd todo
$ make
```
Then todo will be running at: http://localhost:8000

By default todo stores todos in `todo.db` in the local directory.

This can be configured with the `-dbpath=/path/to/todo.db` option.

You can pass in the other environment variables using the flag notation as well, for example:
```
$ todo -maxitems=20 -maxtitlelength=50 -theme=nord
```

## License
MIT

Icon made by [Smashicons](https://smashicons.com/) from [flaticon.com](https://flaticon.com)
