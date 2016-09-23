package store

func StoreProvider() Store {
	return &QiniuStore{}
}
