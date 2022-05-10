# The 2022 Philippine Election Data

A json response copy of `https://www.gmanetwork.com/news/eleksyon2022/`, specifically `e22vh.gmanetwork.com` json api.

Written in simple http request using [Go Language](/crawlgma.go) with goroutine!

## License

I don't own the data lol, thanks to GMA for implementing such detailed json response with Interactive UI, salute to their developers!

Use this whatever you want, maybe create a youtube animation base on the data per batch?

## JSON Format

- `election_returns_processed`
- `location_code`
- `result`
  - index
    - `candidates`
      - index
        - `name` - name of the candidate, e.g: Bong Bong Marcos
        - `party` - e.g: PDP
        - `vote_count` numeric count of the current batch / hourly
    - `contest`
- `result_as_of` - `<datetime>` The date format is "YYYY/MM/DD HH:MM:SS", timezone is GMT+8
- `total_voters_processed`
