# graphql

[![Build Status](https://travis-ci.org/c7/graphql.svg?branch=master)](https://travis-ci.org/c7/graphql)
[![Go Report Card](https://goreportcard.com/badge/github.com/c7/graphql?style=flat)](https://goreportcard.com/report/github.com/c7/graphql)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/c7/graphql)
[![License Apache](https://img.shields.io/badge/license-Apache-lightgrey.svg?style=flat)](https://github.com/c7/graphql#license-apache)

A GraphQL client with no third party dependencies.

Initial version based on <https://github.com/machinebox/graphql>

## Installation

```
go get github.com/c7/graphql
```

## Usage

```go
package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/c7/graphql"
)

func main() {
	gql := graphql.NewClient("https://swapi-graphql.netlify.app/.netlify/functions/index")

	ctx := context.Background()

	req := graphql.NewRequest(`
		query {
			allStarships(first: 4) {
    		totalCount
    		starships {
      		name
      		model
      		crew
    		}
  		}
		}
	`)

	var resp struct {
		AllStarships struct {
			TotalCount int
			Starships  []struct {
				Name  string
				Model string
				Crew  string
			}
		}
	}

	if err := gql.Run(ctx, req, &resp); err != nil {
		panic(err)
	}

	json.NewEncoder(os.Stdout).Encode(resp)
}
```

```json
{
  "AllStarships": {
    "TotalCount": 36,
    "Starships": [
      {
        "Name": "CR90 corvette",
        "Model": "CR90 corvette",
        "Crew": "30-165"
      },
      {
        "Name": "Star Destroyer",
        "Model": "Imperial I-class Star Destroyer",
        "Crew": "47,060"
      },
      {
        "Name": "Sentinel-class landing craft",
        "Model": "Sentinel-class landing craft",
        "Crew": "5"
      },
      {
        "Name": "Death Star",
        "Model": "DS-1 Orbital Battle Station",
        "Crew": "342,953"
      }
    ]
  }
}
```

<img src="https://assets.c7.se/svg/viking-gopher.svg" align="right" width="30%" height="300">

## License (Apache)

Copyright 2020-2023 [Code7 Interactive](https://c7.se)

> Licensed under the Apache License, Version 2.0 (the "License");
> you may not use this file except in compliance with the License.
> You may obtain a copy of the License at
>
>   http://www.apache.org/licenses/LICENSE-2.0
>
> Unless required by applicable law or agreed to in writing, software
> distributed under the License is distributed on an "AS IS" BASIS,
> WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
> See the License for the specific language governing permissions and
> limitations under the License.
