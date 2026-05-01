package db

func (d *sqlDB) CreateUser(username, passwordHash string, initialBalance float64) (int, error) {
	result, err := d.conn.Exec(
		"INSERT INTO users (username, password_hash, balance) VALUES (?, ?, ?)",
		username, passwordHash, initialBalance,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (d *sqlDB) GetUserByUsername(username string) (*User, error) {
	var u User
	err := d.conn.QueryRow(
		"SELECT id, username, password_hash, balance, created_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Balance, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (d *sqlDB) GetUserByID(id int) (*User, error) {
	var u User
	err := d.conn.QueryRow(
		"SELECT id, username, password_hash, balance, created_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Balance, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (d *sqlDB) AddBalance(userID int, amount float64) error {
	_, err := d.conn.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", amount, userID)
	return err
}

func (d *sqlDB) DeductBalance(userID int, amount float64) error {
	_, err := d.conn.Exec("UPDATE users SET balance = balance - ? WHERE id = ?", amount, userID)
	return err
}

func (d *sqlDB) ListUsers() ([]User, error) {
	rows, err := d.conn.Query("SELECT id, username, password_hash, balance, created_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Balance, &u.CreatedAt); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users, nil
}
