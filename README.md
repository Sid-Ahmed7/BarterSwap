# BarterSwap — API 

API REST en Go.

## Prérequis

- Docker

## Installation

```bash
git clone <url>
cd projet-BarterSwap
docker compose up --build
```


## Variables d'environnement

Définies dans le fichier `.env` à la racine :

| Variable          | Description              |
|-------------------|--------------------------|
| POSTGRES_USER     | Utilisateur PostgreSQL   |
| POSTGRES_PASSWORD | Mot de passe PostgreSQL  |
| POSTGRES_DB       | Nom de la base           |
| DB_HOST           | Hôte de la base          |
| DB_PORT           | Port PostgreSQL           |
| SERVER_PORT       | Port du serveur HTTP     |
