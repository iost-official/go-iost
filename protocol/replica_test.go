package protocol_test

//func TestReplica(t *testing.T) {
//	Convey("Test of Replica", t, func() {
//		rep, _ := ReplicaFactory("pbft")
//		Convey("Init, run and stop", func() {
//			mock := NewController(t)
//			defer mock.Finish()
//
//			m, _ := iosbase.NewMember(nil)
//			db := NewMockDatabase(mock)
//			router := NewMockRouter(mock)
//
//			chv := make(chan View)
//			chreq := make(chan iosbase.Request)
//			chres := make(chan iosbase.Response)
//			db.EXPECT().NewViewSignal().AnyTimes().Return(chv, nil)
//			router.EXPECT().FilteredChan(Any()).AnyTimes().Return(chreq, chres, nil)
//
//			err := rep.Init(m, db, router)
//
//			So(err, ShouldBeNil)
//
//			var wg sync.WaitGroup
//			wg.Add(1)
//			go func() {
//				rep.Run()
//				wg.Done()
//			}()
//			time.Sleep(10 * time.Millisecond)
//			rep.Stop()
//			wg.Wait()
//			So(true, ShouldBeTrue)
//		})
//	})
//}
