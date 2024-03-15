# gotennis
Tennis simulation written in golang. Can be used to simulate a tennis match between 2 players.

## Requirements

Install [Go](https://golang.org), developed with version 1.21. Should work with older versions.

## Installation
Use got get
```
$ go get github.com/ttopias/gotennis
```
Then import the package into your own code:
```
import "github.com/ttopias/gotennis"
```

## Usage
Figure out serve and return probabilities for both of the players in a match **against** the opponent. Then call:

```
result, err := gotennis.SimulateMatch(playerA, playerB, 100000, 5)
```

Result contains a simulation result after 100k simulations for a best-of 5 tennis match.

## Contribute
Use issues for everything

- Report problems
- Suggest new features
- Improve/fix documentation

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