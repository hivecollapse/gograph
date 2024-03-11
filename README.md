# gograph

A command line tool to explore and test graphql endpoints.

- Generate graphql query text from a schema
- Generate graphql query input stub from schema
- Run sequence of graphql queries and perform basic validation on the output

# Install

Download the release and put the binary anywhere.

You can install the lastest release directly to /usr/local/bin using an automatically generated install script from https://github.com/jpillora/installer

```sh
curl https://i.jpillora.com/hivecollapse/gograph! | bash
```

⚠️ The script is auto-generated, you should inspect it before executing it.

# Usage

## `gograph --help`

Print help

## `gograph schema query ls`

List queries in a schema

### Example

```sh
gograph schema --path "sample/starwars/*.graphql" query ls
```

**Output**

```txt
query:  allFilms(after: String, first: Int, before: String, last: Int): FilmsConnection
query:  film(id: ID, filmID: ID): Film
query:  allPeople(after: String, first: Int, before: String, last: Int): PeopleConnection
query:  person(id: ID, personID: ID): Person
query:  allPlanets(after: String, first: Int, before: String, last: Int): PlanetsConnection
query:  planet(id: ID, planetID: ID): Planet
query:  allSpecies(after: String, first: Int, before: String, last: Int): SpeciesConnection
query:  species(id: ID, speciesID: ID): Species
query:  allStarships(after: String, first: Int, before: String, last: Int): StarshipsConnection
[...]
```

## `gograph schema query gen <operation>`

Generate query text for a graphql operation.

### Arguments

`--schema <glob>` Path to the graphql schema. Glob are accepted and multiple files are merged into a single schema.  
`--level <number>` Change the depth of the query. The default depth is set to 3.

### Example

Generate the query for the list of `allFilms` in the starwars api.

```sh
gograph schema --path "sample/starwars/*.graphql" query gen allFilms -l1
```

**Sample output**

```graphql
query AllFilms($after: String, $first: Int, $before: String, $last: Int) {
	allFilms(after: $after, first: $first, before: $before, last: $last) {
		pageInfo {
			hasNextPage
			hasPreviousPage
			startCursor
			endCursor
		}
		edges {
			cursor
		}
		totalCount
		films {
			title
			episodeID
			openingCrawl
			director
			producers
			releaseDate
			created
			edited
			id
		}
	}
}
```

## `gograph schema query genvar <operation>`

Generate graphql query input stub from schema

### Example

Generate the input variables for the `dragon` spacex api.

```sh
gograph schema --path "sample/starwars/*.graphql" query genvar allFilms;
```

**Sample output**

```json
{
	"after": null,
	"before": null,
	"first": null,
	"last": null
}
```

## `gograph flow <flow.yml>,...`

Execute a list of grapqhl query defined in a flow file.

**Example**

```sh
gograph flow run sample/starwars/flow.yml
```

See [sample/starwars/flow.yml](starwars sample flow) for an example.
