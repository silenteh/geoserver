package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

type adminPanel struct {
	db *bolt.DB
}

func NewAdminPanel(db *bolt.DB) *adminPanel {

	return &adminPanel{
		db: db,
	}
}

func bucket() []byte {
	return []byte("vms")
}

func (s *adminPanel) ping(z *zone) error {

	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket())
		if err != nil {
			return fmt.Errorf("Error creating bucket: %s", err)
		}
		serialized := z.toJson()

		if len(serialized) > 0 {
			log.Printf("Zone %s stored correctly\n", serialized)
			return b.Put([]byte(z.IpAddress.Ip), serialized)
		}

		return nil
	})

}

func (s *adminPanel) getAll(vms *[]*zone) error {
	return s.db.View(func(tx *bolt.Tx) error {

		var allVms []*zone

		// Assume bucket exists and has keys
		b := tx.Bucket(bucket())

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			t, err := ZoneFromJson(v)
			if err != nil {
				log.Println("Could not deserialize zone", err)
				return err
			}

			allVms = append(allVms, t)
			*vms = allVms
		}

		return nil
	})
}
