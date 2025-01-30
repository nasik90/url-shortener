package storage

import "testing"

func TestLocalCache_RestoreData(t *testing.T) {
	type args struct {
		filePath string
	}
	// cache := make(map[string]string)
	tests := []struct {
		name       string
		localCache *LocalCache
		args       args
		wantErr    bool
	}{
		{
			name:       "success test",
			localCache: &LocalCache{CahceMap: make(map[string]string)},
			args:       args{filePath: "C:/_temp/store.txt"},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.localCache.RestoreData(tt.args.filePath); (err != nil) != tt.wantErr {
				t.Errorf("LocalCache.RestoreData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
