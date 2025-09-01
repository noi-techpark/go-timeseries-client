<!--
SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>

SPDX-License-Identifier: CC0-1.0
-->

# Golang client for Open Data Hub time series API (affectionately named "ninja")

This is a golang client library for the Open Data Hub time series API:
[Ninja API](https://github.com/noi-techpark/it.bz.opendatahub.api.mobility-ninja)

## Usage
`go get github.com/noi-techpark/go-timeseries-client`


```go
	// a short string identifying your application
	referer := "myapp.com"
	// create a client
	c := odhts.NewDefaultClient(referer)
	// if you want to use a different Timeseries or Authentication endpoints, create a custom client instead
	customClient := odhts.NewCustomClient(
		"http://localhost:8080",
		"http://authserver.example.com/auth/realms/test/protocol/openid-connect/token",
		referer)
	// (optional) use oauth credentials for requests
	customClient.UseAuth("myclientid", "myclientsecret")

	// Prepare a new request
	// Refer to the OpenAPI spec at https://mobility.api.opendatahub.com for all options
	req := odhts.DefaultRequest()
	req.Repr = odhts.FlatNode
	req.AddStationType("EChargingStation")
	req.AddDataType("number-available")
	// add some custom filters
	req.Where = where.And(
		where.Eq("sorigin", "Neogy"),
		where.Eq("sactive", "true"),
	)
	req.Limit = 5

	// request stations
	// We use a provided response Dto, but you can (and sometimes have to) pass your own JSON-mappable types
	var stations odhts.Response[[]odhts.StationDto[map[string]any]]
	if err := odhts.StationType(c, req, &stations); err != nil {
		panic(err)
	}
	fmt.Printf("Stations:\n %v\n\n", stations)

	// request latest measurements
	var latest odhts.Response[[]odhts.LatestDto]
	if err := odhts.Latest(c, req, &latest); err != nil {
		panic(err)
	}
	fmt.Printf("Measurements:\n %v", latest)

```

## Information

### Support

For support, please contact [help@opendatahub.com](mailto:help@opendatahub.com).

### Contributing

If you'd like to contribute, please follow our [Getting
Started](https://github.com/noi-techpark/odh-docs/wiki/Contributor-Guidelines:-Getting-started)
instructions.
### License
The code in this project is licensed under Mozilla Public License Version 2.0

### REUSE

This project is [REUSE](https://reuse.software) compliant, more information about the usage of REUSE in NOI Techpark repositories can be found [here](https://github.com/noi-techpark/odh-docs/wiki/Guidelines-for-developers-and-licenses#guidelines-for-contributors-and-new-developers).

Since the CI for this project checks for REUSE compliance you might find it useful to use a pre-commit hook checking for REUSE compliance locally. The [pre-commit-config](.pre-commit-config.yaml) file in the repository root is already configured to check for REUSE compliance with help of the [pre-commit](https://pre-commit.com) tool.

Install the tool by running:
```bash
pip install pre-commit
```
Then install the pre-commit hook via the config file by running:
```bash
pre-commit install
```