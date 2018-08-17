package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func ravendb_8761_can_group_by_array_values(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	ravendb_8761_putDocs(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Advanced().RawQuery(ravendb.GetTypeOf(&ProductCount{}), "from Orders group by lines[].product\n"+
			"  order by count()\n"+
			"  select key() as productName, count() as count")
		q = q.WaitForNonStaleResults()
		productCounts1, err := q.ToList()
		assert.NoError(t, err)

		q2 := session.Advanced().DocumentQuery(ravendb.GetTypeOf(&Order{}))
		q3 := q2.GroupBy("lines[].product")
		q3 = q3.SelectKeyWithNameAndProjectedName("", "productName")
		q2 = q3.SelectCount()
		q2 = q2.OfType(ravendb.GetTypeOf(&ProductCount{}))
		productCounts2, err := q2.ToList()
		assert.NoError(t, err)

		combined := [][]interface{}{productCounts1, productCounts2}
		for _, products := range combined {
			assert.Equal(t, len(products), 2)

			product := products[0].(*ProductCount)
			assert.Equal(t, product.ProductName, "products/1")
			assert.Equal(t, product.Count, 1)

			product = products[1].(*ProductCount)
			assert.Equal(t, product.ProductName, "products/2")
			assert.Equal(t, product.Count, 2)
		}
		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q := session.Advanced().RawQuery(ravendb.GetTypeOf(&ProductCount{}), "from Orders\n"+
			" group by lines[].product, shipTo.country\n"+
			" order by count() \n"+
			" select lines[].product as productName, shipTo.country as country, count() as count")
		productCounts1, err := q.ToList()
		assert.NoError(t, err)

		q2 := session.Advanced().DocumentQuery(ravendb.GetTypeOf(&Order{}))
		q3 := q2.GroupBy("lines[].product", "shipTo.country")
		q3 = q3.SelectKeyWithNameAndProjectedName("lines[].product", "productName")
		q3 = q3.SelectKeyWithNameAndProjectedName("shipTo.country", "country")
		q2 = q3.SelectCount()
		q2 = q2.OfType(ravendb.GetTypeOf(&ProductCount{}))
		productCounts2, err := q2.ToList()
		assert.NoError(t, err)

		combined := [][]interface{}{productCounts1, productCounts2}
		for _, products := range combined {
			assert.Equal(t, len(products), 2)

			product := products[0].(*ProductCount)
			assert.Equal(t, product.ProductName, "products/1")
			assert.Equal(t, product.Count, 1)
			assert.Equal(t, product.Country, "USA")

			product = products[1].(*ProductCount)
			assert.Equal(t, product.ProductName, "products/2")
			assert.Equal(t, product.Count, 2)
			assert.Equal(t, product.Country, "USA")
		}
		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q := session.Advanced().RawQuery(ravendb.GetTypeOf(&ProductCount{}), "from Orders\n"+
			" group by lines[].product, lines[].quantity\n"+
			" order by lines[].quantity\n"+
			" select lines[].product as productName, lines[].quantity as quantity, count() as count")
		productCounts1, err := q.ToList()
		assert.NoError(t, err)

		q2 := session.Advanced().DocumentQuery(ravendb.GetTypeOf(&Order{}))
		q3 := q2.GroupBy("lines[].product", "lines[].quantity")
		q3 = q3.SelectKeyWithNameAndProjectedName("lines[].product", "productName")
		q3 = q3.SelectKeyWithNameAndProjectedName("lines[].quantity", "quantity")
		q2 = q3.SelectCount()
		q2 = q2.OfType(ravendb.GetTypeOf(&ProductCount{}))
		productCounts2, err := q2.ToList()

		combined := [][]interface{}{productCounts1, productCounts2}
		for _, products := range combined {
			assert.Equal(t, len(products), 3)

			product := products[0].(*ProductCount)
			assert.Equal(t, product.ProductName, "products/1")

			assert.Equal(t, product.Count, 1)
			assert.Equal(t, product.Quantity, 1)

			product = products[1].(*ProductCount)
			assert.Equal(t, product.ProductName, "products/2")
			assert.Equal(t, product.Count, 1)
			assert.Equal(t, product.Quantity, 2)

			product = products[2].(*ProductCount)
			assert.Equal(t, product.ProductName, "products/2")
			assert.Equal(t, product.Count, 1)
			assert.Equal(t, product.Quantity, 3)
		}
		session.Close()
	}
}

func ravendb_8761_can_group_by_array_content(t *testing.T) {

}

type ProductCount struct {
	ProductName string   `json:"productName"`
	Count       int      `json:"count"`
	Country     string   `json:"country"`
	Quantity    int      `json:"quantity"`
	Products    []string `json:"products"`
	Quantities  []int    `json:"quantities"`
}

func ravendb_8761_putDocs(t *testing.T, store *ravendb.IDocumentStore) {
	var err error

	session := openSessionMust(t, store)
	order1 := &Order{}

	orderLine11 := &OrderLine{
		Product:  "products/1",
		Quantity: 1,
	}

	orderLine12 := &OrderLine{
		Product:  "products/2",
		Quantity: 2,
	}

	order1.Lines = []*OrderLine{orderLine11, orderLine12}

	address1 := &Address{
		Country: "USA",
	}

	order1.ShipTo = address1

	err = session.Store(order1)
	assert.NoError(t, err)

	orderLine21 := &OrderLine{
		Product:  "products/2",
		Quantity: 3,
	}

	address2 := &Address{
		Country: "USA",
	}
	order2 := &Order{
		Lines:  []*OrderLine{orderLine21},
		ShipTo: address2,
	}

	err = session.Store(order2)
	assert.NoError(t, err)

	err = session.SaveChanges()
	assert.NoError(t, err)

	session.Close()
}

func TestRavenDB8761(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches the order of Java tests
	ravendb_8761_can_group_by_array_content(t)
	ravendb_8761_can_group_by_array_values(t)
}
