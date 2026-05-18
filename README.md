# Internship Tracker

The personal infinity stone gauntlet for college students everywhere: an internship tracker.

I originally set this up as a SvelteKit application (because Svelte is cool!) and had ChatGPT shove in all the code, but I've come to my senses: why have many file when few file do trick? The SvelteKit skeleton drops so many files and spread them literally everywhere. Go applications are so much easier to reason about than Sveltekit because with Go, all you need is one `main.go` (plus templates) and you're good to... Go. Haha.

So I shoved the old SvelteKit application into ChatGPT and had it spit out Go and HTML templates. And now it starts up instantly unlike Svelte! It's so fast, even switching pages are fast! Weird how locally-accessed pages *weren't fast* before. But Go made everything faster. That's the power of the old Internet right there, renamed to so-called "server-side rendering" now. Amazing, how web technologies are like a pendulum.

## Requirements

- Go 1.23+ with cgo (sadly!).

## Run

```sh
export DATABASE_URL=./internship-tracker.sqlite
export ADDR=:5173
go run .
```

Then open `http://localhost:5173`.

If `DATABASE_URL` is not set, the app defaults to `./internship-tracker.sqlite`.

## Routes

- `/` dashboard
- `/add` add application
- `/app/{id}` edit application and timeline
- `/calendar.ics` interview calendar feed

## Build

```sh
go build -o internship-tracker .
DATABASE_URL=/path/to/existing.sqlite ./internship-tracker
```

## License

The code started out the SvelteKit skeleton template with Drizzle (I don't know what license that's under!), then I stuffed ChatGPT code in then I laundered the entire project into ChatGPt, so basically I gave ChatGPT's code to ChatGPT to generate more ChatGPT code. Professional vibecoder right here. I don't think I have any copyright over what's in this repository, but let's just pretend the license below applies, alright?

This software is licensed under the MIT License. See [LICENSE](https://github.com/ganyuke/internship-tracker/blob/master/LICENSE) for details.
