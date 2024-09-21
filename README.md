# `tmap`

A `TMap` is a wrapper around `sync.Map` with context and a TTL.  
It has the same methods as `sync.Map` plus `Truncate` and `Flush`, which takes care of removing all items in a `Map` and removing expired items from a `Map`, respectivelly.

## TTL 

The idea behind this is to have a toy cache for very specific situations. In case you never call `go tmap.Flush`, these items will never be removed unless you call `tmap.Truncate`, so in a way it's flexible.

Particularly, you'll want to use a low expiration time (ttl) and a _lower_ ticker, this way you'll never retrieve expired data. Since this code is fairly simple, it does not checks if the data loaded is expired, thats what `tmap.Flush` is for.


## Example

There's a sort of real world example in [example/](example/) but this snippet shows pretty much how it can be used. Also, you can take a look at the tests (which is covers this lib by a 100%). 


```
type example struct {
	K string
	V string
}

func (e example) GetID() string {
	return e.K
}

// ...

m := tmap.NewTMap(10*time.Second, time.Now)
t := time.NewTicker(1 * time.Second)
go m.Flush(ctx, t)

if err := m.Store(ctx, example{"example-key-1", "example content}); err != nil {
    fmt.Println("something went wrong:", err)
    return
}

value, err := m.Load(ctx, "example-key-1")
if err != nil {
    fmt.Println("something went wrong:", err)
    return
}

fmt.Printf("%+v\n", value)
```

## License 

See [LICENSE](LICENSE)

