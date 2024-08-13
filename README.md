# sqidsencoder

`sqidsencoder` is a Go library that provides functionality to encode and decode structs using [sqids](https://github.com/sqids/sqids-go)

## Usage

### Encoding

To encode a struct:

```go
func (enc sqidsencoder) Encode(src any, dst any) error
```

- `src`: The source struct to be encoded.
- `dst`: A pointer to the destination struct where the encoded values will be stored.

### Decoding

To decode a struct:

```go
func (enc sqidsencoder) Decode(src any, dst any) error
```

- `src`: The source struct containing encoded values.
- `dst`: A pointer to the destination struct where the decoded values will be stored.

## Struct Tags

The library uses struct tags to determine which fields should be encoded or decoded. The tag key is `sqids`.

- Use `sqids:"encode"` to mark a field for encoding.
- Use `sqids:"decode"` to mark a field for decoding.

## Example

### Simple example
```go
package main

import (
	"fmt"
	"log"

	"github.com/andersonjoseph/sqidsencoder"
	"github.com/sqids/sqids-go"
)

type User struct {
    ID   uint64    `sqids:"encode"`
    Name string
}

type EncodedUser struct {
    ID   string
    Name string
}

func main() {
    s, err := sqids.New()

    if err != nil {
        log.Fatal(err)
    }

    encoder := sqidsencoder.New(s)

    user := User{ ID: 123, Name: "John Doe" }
    var encodedUser EncodedUser

    err = encoder.Encode(user, &encodedUser)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("plan ID: %v\n", user.ID) // plan ID: 123
    fmt.Printf("encoded ID: %v\n", encodedUser.ID) // encoded ID: Ukk
}
```

### Advanced example
```go
type Order struct {
	ID   uint64 `sqids:"encode"`

    // nested struct
	User struct {
		ID   uint64 `sqids:"encode"`
		Name string
	} `sqids:"encode"`

    // slice
	GiftCards []uint64 `sqids:"encode"`

    // slice of struct
	Items []struct {
		ID   uint64 `sqids:"encode"`
		Name string
	} `sqids:"encode"`
}

type EncodedOrder struct {
	ID   string
	User struct {
		ID   string
		Name string
	}
	Items []struct {
		ID   string
		Name string
	}
	GiftCards []string
}

func main() {
	s, err := sqids.New()
	if err != nil {
		log.Fatal(err)
	}
	encoder := sqidsencoder.New(s)

	order := Order{
		ID: 1021,
		User: struct {
			ID   uint64 `sqids:"encode"`
			Name string
		}{
			ID:   986,
			Name: "Jhon Doe",
		},
		Items: []struct {
			ID   uint64 `sqids:"encode"`
			Name string
		}{
			{ID: 293, Name: "cool item A"},
			{ID: 993, Name: "cool item B"},
		},
		GiftCards: []uint64{5, 6, 7, 8},
	}

	var encodedOrder EncodedOrder

	err = encoder.Encode(order, &encodedOrder)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("encodedOrder: %v\n", encodedOrder)
}
```

## Error Handling

| Error Name | Returned When |
|------------|---------------|
| `ErrInvalidInput` | - Incorrect data structures provided<br>- Missing required fields<br>- Invalid field values |
| `ErrType` | - Type conversion failures occur<br>- Unexpected types in encoding/decoding operations<br>- Incompatible types in assignments |
| `ErrInvalidID` | - Malformed IDs<br>- IDs don't match expected patterns<br>|

### Example

```go
err := someFunction()
switch {
case errors.Is(err, ErrInvalidInput):
    // Handle invalid input
case errors.Is(err, ErrType):
    // Handle type mismatch
case errors.Is(err, ErrInvalidID):
    // Handle invalid ID
default:
    // Handle other errors
}
