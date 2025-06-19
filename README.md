# gotennis

Tennis match simulation service written in Go. Simulates tennis matches between two players and provides probabilities for various betting markets via a simple HTTP API.

## Features

- Simulate tennis matches between two players
- Supports best-of-3 and best-of-5 formats
- Configurable number of simulations (default: 1,000,000)
- Returns probabilities for moneyline, set/game handicaps, and totals
- Runs as a standalone HTTP service (Docker or native)

## API Usage

Send a GET request to the root endpoint `/` with the following query parameters:

- `p1`: Probability of player 1 winning a point on serve (float, required)
- `p2`: Probability of player 2 winning a point on serve (float, required)
- `bestof`: Number of sets (3 or 5, required)
- `simulations`: Number of simulations to run (optional, default: 1,000,000)

Example:

```sh
curl "http://localhost:8000/?p1=0.65&p2=0.60&bestof=5&simulations=500000"
```

## Running as a Docker Service

1. **Build the Docker image:**

   ```sh
   docker build -t gotennis .
   ```

2. **Run the service:**

   ```sh
   docker run -p 8000:8000 -e GOTENNIS_PORT=8000 gotennis
   ```

   The service will listen on the port defined by the `GOTENNIS_PORT` environment variable (default: 8000).

## Running Locally (without Docker)

```sh
export GOTENNIS_PORT=8000
make run
```

## Match Game Simulation Formula

The probability of a player winning a game on serve is calculated as:

```text
P(win from deuce) = p^2 / (1 - 2*p*(1-p))
```

Where `p` is the probability of winning a point on serve.

The overall probability of winning a game is:

```text
P(game) = P(4-0) + P(4-1) + P(4-2) + P(reach deuce) * P(win from deuce)
```

Where:

- `P(4-0) = p^4`
- `P(4-1) = 4 * p^4 * (1-p)`
- `P(4-2) = 10 * p^4 * (1-p)^2`
- `P(reach deuce) = 20 * p^3 * (1-p)^3`

## Statistics Endpoint

The API provides a `/stats` endpoint to retrieve recent request statistics and performance metrics.

- **Endpoint:** `GET /stats`
- **Response:** JSON object with total requests, success/error counts, and averages for simulations, simulation time, and response time (over the last 1000 requests).

Example:

```sh
curl "http://localhost:8000/stats"
```

Sample response:

```json
{
  "total_requests": 42,
  "success_count": 40,
  "error_count": 2,
  "avg_simulations": 1000000,
  "avg_simulation_time_ms": 120.5,
  "avg_response_time_ms": 125.3
}
```

## License

MIT License

Copyright (c) 2024 [ttopias](https://github.com/ttopias)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
