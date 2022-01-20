package access_control_test

//
//func TestShareDocument(t *testing.T) {
//	dbc := db.ConnectForTesting(t)
//	ctx := context.Background()
//
//	userID, err := users.Register(ctx, dbc, "user", "password")
//	require.NoError(t, err)
//
//	documents.Upload(ctx, dbc, userID)
//
//	type args struct {
//		doc         documents_models.Document
//		userID      int64
//		shareUserID int64
//		permissions access_control_models.PERM
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := access_control.ShareDocument(ctx, dbc, tt.args.doc, tt.args.userID, tt.args.shareUserID, tt.args.permissions); (err != nil) != tt.wantErr {
//				t.Errorf("ShareDocument() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
