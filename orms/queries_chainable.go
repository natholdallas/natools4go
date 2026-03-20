package orms

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Model specify the model you would like to run db operations
//
//	// update all users's name to `hello`
//	db.Model(&User{}).Update("name", "hello")
//	// if user's primary key is non-blank, will use it as condition, then will only update that user's name to `hello`
//	db.Model(&user).Update("name", "hello")
func (q *Query[T]) Model(value any) *Query[T] {
	q.db = q.db.Model(value)
	return q
}

// Clauses Add clauses
//
// This supports both standard clauses (clause.OrderBy, clause.Limit, clause.Where) and more
// advanced techniques like specifying lock strength and optimizer hints. See the
// [docs] for more depth.
//
//	// add a simple limit clause
//	db.Clauses(clause.Limit{Limit: 1}).Find(&User{})
//	// tell the optimizer to use the `idx_user_name` index
//	db.Clauses(hints.UseIndex("idx_user_name")).Find(&User{})
//	// specify the lock strength to UPDATE
//	db.Clauses(clause.Locking{Strength: "UPDATE"}).Find(&users)
//
// [docs]: https://gorm.io/docs/sql_builder.html#Clauses
func (q *Query[T]) Clauses(conds ...clause.Expression) *Query[T] {
	q.db = q.db.Clauses(conds...)
	return q
}

// Distinct specify distinct fields that you want querying
//
//	// Select distinct names of users
//	db.Distinct("name").Find(&results)
//	// Select distinct name/age pairs from users
//	db.Distinct("name", "age").Find(&results)
func (q *Query[T]) Distinct(args ...any) *Query[T] {
	q.db = q.db.Distinct(args...)
	return q
}

// Select specify fields that you want when querying, creating, updating
//
// Use Select when you only want a subset of the fields. By default, GORM will select all fields.
// Select accepts both string arguments and arrays.
//
//	// Select name and age of user using multiple arguments
//	db.Select("name", "age").Find(&users)
//	// Select name and age of user using an array
//	db.Select([]string{"name", "age"}).Find(&users)
func (q *Query[T]) Select(query any, args ...any) *Query[T] {
	q.db = q.db.Select(query, args...)
	return q
}

// Omit specify fields that you want to ignore when creating, updating and querying
func (q *Query[T]) Omit(columns ...string) *Query[T] {
	q.db = q.db.Omit(columns...)
	return q
}

// MapColumns modify the column names in the query results to facilitate align to the corresponding structural fields
func (q *Query[T]) MapColumns(m map[string]string) *Query[T] {
	q.db = q.db.MapColumns(m)
	return q
}

// Where add conditions
//
// See the [docs] for details on the various formats that where clauses can take. By default, where clauses chain with AND.
//
//	// Find the first user with name jinzhu
//	db.Where("name = ?", "jinzhu").First(&user)
//	// Find the first user with name jinzhu and age 20
//	db.Where(&User{Name: "jinzhu", Age: 20}).First(&user)
//	// Find the first user with name jinzhu and age not equal to 20
//	db.Where("name = ?", "jinzhu").Where("age <> ?", "20").First(&user)
//
// [docs]: https://gorm.io/docs/query.html#Conditions
func (q *Query[T]) Where(query any, args ...any) *Query[T] {
	q.db = q.db.Where(query, args...)
	return q
}

// Not add NOT conditions
//
// Not works similarly to where, and has the same syntax.
//
//	// Find the first user with name not equal to jinzhu
//	db.Not("name = ?", "jinzhu").First(&user)
func (q *Query[T]) Not(query any, args ...any) *Query[T] {
	q.db = q.db.Not(query, args...)
	return q
}

// Or add OR conditions
//
// Or is used to chain together queries with an OR.
//
//	// Find the first user with name equal to jinzhu or john
//	db.Where("name = ?", "jinzhu").Or("name = ?", "john").First(&user)
func (q *Query[T]) Or(query any, args ...any) *Query[T] {
	q.db = q.db.Or(query, args...)
	return q
}

// Joins specify Joins conditions
//
//	db.Joins("Account").Find(&user)
//	db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").Find(&user)
//	db.Joins("Account", DB.Select("id").Where("user_id = users.id AND name = ?", "someName").Model(&Account{}))
func (q *Query[T]) Joins(query string, args ...any) *Query[T] {
	q.db = q.db.Joins(query, args...)
	return q
}

// InnerJoins specify inner joins conditions
// db.InnerJoins("Account").Find(&user)
func (q *Query[T]) InnerJoins(query string, args ...any) *Query[T] {
	q.db = q.db.InnerJoins(query, args...)
	return q
}

// Group specify the group method on the find
//
//	// Select the sum age of users with given names
//	db.Model(&User{}).Select("name, sum(age) as total").Group("name").Find(&results)
func (q *Query[T]) Group(name string) *Query[T] {
	q.db = q.db.Group(name)
	return q
}

// Having specify HAVING conditions for GROUP BY
//
//	// Select the sum age of users with name jinzhu
//	db.Model(&User{}).Select("name, sum(age) as total").Group("name").Having("name = ?", "jinzhu").Find(&result)
func (q *Query[T]) Having(query any, args ...any) *Query[T] {
	q.db = q.db.Having(query, args...)
	return q
}

// Order specify order when retrieving records from database
//
//	db.Order("name DESC")
//	db.Order(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: true})
//	db.Order(clause.OrderBy{Columns: []clause.OrderByColumn{
//		{Column: clause.Column{Name: "name"}, Desc: true},
//		{Column: clause.Column{Name: "age"}, Desc: true},
//	}})
func (q *Query[T]) Order(value any) *Query[T] {
	q.db = q.db.Order(value)
	return q
}

// Limit specify the number of records to be retrieved
//
// Limit conditions can be cancelled by using `Limit(-1)`.
//
//	// retrieve 3 users
//	db.Limit(3).Find(&users)
//	// retrieve 3 users into users1, and all users into users2
//	db.Limit(3).Find(&users1).Limit(-1).Find(&users2)
func (q *Query[T]) Limit(limit int) *Query[T] {
	q.db = q.db.Limit(limit)
	return q
}

// Offset specify the number of records to skip before starting to return the records
//
// Offset conditions can be cancelled by using `Offset(-1)`.
//
//	// select the third user
//	db.Offset(2).First(&user)
//	// select the first user by cancelling an earlier chained offset
//	db.Offset(5).Offset(-1).First(&user)
func (q *Query[T]) Offset(offset int) *Query[T] {
	q.db = q.db.Offset(offset)
	return q
}

// Scopes pass current database connection to arguments `func(DB) DB`, which could be used to add conditions dynamically
//
//	func AmountGreaterThan1000(db *gorm.DB) *gorm.DB {
//	    return db.Where("amount > ?", 1000)
//	}
//
//	func OrderStatus(status []string) func (db *gorm.DB) *gorm.DB {
//	    return func (db *gorm.DB) *gorm.DB {
//	        return db.Scopes(AmountGreaterThan1000).Where("status in (?)", status)
//	    }
//	}
//
//	db.Scopes(AmountGreaterThan1000, OrderStatus([]string{"paid", "shipped"})).Find(&orders)
func (q *Query[T]) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *Query[T] {
	q.db = q.db.Scopes(funcs...)
	return q
}

// Preload preload associations with given conditions
//
//	// get all users, and preload all non-cancelled orders
//	db.Preload("Orders", "state NOT IN (?)", "cancelled").Find(&users)
func (q *Query[T]) Preload(query string, args ...any) *Query[T] {
	q.db = q.db.Preload(query, args...)
	return q
}

// Attrs provide attributes used in [FirstOrCreate] or [FirstOrInit]
//
// Attrs only adds attributes if the record is not found.
//
//	// assign an email if the record is not found
//	db.Where(User{Name: "non_existing"}).Attrs(User{Email: "fake@fake.org"}).FirstOrInit(&user)
//	// user -> User{Name: "non_existing", Email: "fake@fake.org"}
//
//	// assign an email if the record is not found, otherwise ignore provided email
//	db.Where(User{Name: "jinzhu"}).Attrs(User{Email: "fake@fake.org"}).FirstOrInit(&user)
//	// user -> User{Name: "jinzhu", Age: 20}
//
// [FirstOrCreate]: https://gorm.io/docs/advanced_query.html#FirstOrCreate
// [FirstOrInit]: https://gorm.io/docs/advanced_query.html#FirstOrInit
func (q *Query[T]) Attrs(attrs ...any) *Query[T] {
	q.db = q.db.Attrs(attrs...)
	return q
}

// Assign provide attributes used in [FirstOrCreate] or [FirstOrInit]
//
// Assign adds attributes even if the record is found. If using FirstOrCreate, this means that
// records will be updated even if they are found.
//
//	// assign an email regardless of if the record is not found
//	db.Where(User{Name: "non_existing"}).Assign(User{Email: "fake@fake.org"}).FirstOrInit(&user)
//	// user -> User{Name: "non_existing", Email: "fake@fake.org"}
//
//	// assign email regardless of if record is found
//	db.Where(User{Name: "jinzhu"}).Assign(User{Email: "fake@fake.org"}).FirstOrInit(&user)
//	// user -> User{Name: "jinzhu", Age: 20, Email: "fake@fake.org"}
//
// [FirstOrCreate]: https://gorm.io/docs/advanced_query.html#FirstOrCreate
// [FirstOrInit]: https://gorm.io/docs/advanced_query.html#FirstOrInit
func (q *Query[T]) Assign(attrs ...any) *Query[T] {
	q.db = q.db.Assign(attrs...)
	return q
}

// Unscoped disables the global scope of soft deletion in a query.
// By default, GORM uses soft deletion, marking records as "deleted"
// by setting a timestamp on a specific field (e.g., `deleted_at`).
// Unscoped allows queries to include records marked as deleted,
// overriding the soft deletion behavior.
// Example:
//
//	var users []User
//	db.Unscoped().Find(&users)
//	// Retrieves all users, including deleted ones.
func (q *Query[T]) Unscoped() *Query[T] {
	q.db = q.db.Unscoped()
	return q
}

// Raw is a raw SQL builder with placeholders.
func (q *Query[T]) Raw(sql string, values ...any) *Query[T] {
	q.db = q.db.Raw(sql, values...)
	return q
}
