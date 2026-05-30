CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    pseudo VARCHAR(100) NOT NULL UNIQUE,
    bio TEXT,
    ville VARCHAR(100),
    credit_balance INT NOT NULL DEFAULT 10,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS skills (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    nom VARCHAR(100) NOT NULL,
    niveau VARCHAR(20) NOT NULL CHECK (niveau IN ('débutant', 'intermédiaire', 'expert')),
    PRIMARY KEY (user_id, nom)
);

CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    provider_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    titre VARCHAR(200) NOT NULL,
    description TEXT,
    categorie VARCHAR(50) NOT NULL CHECK (categorie IN (
                        'Informatique', 'Jardinage', 'Bricolage', 'Cuisine', 'Musique',
                        'Langues', 'Sport', 'Tutorat', 'Déménagement', 'Photographie',
                        'Animalier', 'Couture', 'Autre'
                    )),
    duree_minutes INT NOT NULL,
    credits INT NOT NULL,
    ville VARCHAR(100),
    actif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS exchanges (
    id SERIAL PRIMARY KEY,
    service_id INT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    requester_id INT NOT NULL REFERENCES users(id),
    owner_id INT NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'cancelled', 'completed')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS credit_transactions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    exchange_id INT NOT NULL REFERENCES exchanges(id),
    montant INT NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('earn', 'spend', 'refund')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    exchange_id  INT NOT NULL REFERENCES exchanges(id),
    author_id INT NOT NULL REFERENCES users(id),
    target_id INT NOT NULL REFERENCES users(id),
    note INT NOT NULL CHECK (note BETWEEN 1 AND 5),
    commentaire TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (exchange_id, author_id)
);
