# BarterSwap — API d'échange de compétences

API REST en Go permettant l'échange de compétences entre particuliers via un système de crédits-temps. Chaque heure de service rendue donne droit à une heure de service reçue.

## Prérequis

- Docker

## Installation

```bash
git clone <url>
cd projet-BarterSwap
docker compose up --build
```

L'API est disponible sur `http://localhost:8080`.
La documentation interactive Swagger est disponible sur `http://localhost:8080/swagger/index.html`.

## Variables d'environnement

Définies dans le fichier `.env` à la racine :

| Variable          | Description              |
|-------------------|--------------------------|
| POSTGRES_USER     | Utilisateur PostgreSQL   |
| POSTGRES_PASSWORD | Mot de passe PostgreSQL  |
| POSTGRES_DB       | Nom de la base           |
| DB_HOST           | Hôte de la base          | 
| DB_PORT           | Port PostgreSQL interne  |
| DB_EXPOSED_PORT   | Port PostgreSQL exposé sur l'hôte |
| SERVER_PORT       | Port du serveur HTTP     | 

> L'authentification utilise le header `X-User-ID` (ID de l'utilisateur connecté).

## Endpoints

### Utilisateurs

| Méthode | Path                       | Description                              |
|---------|----------------------------|------------------------------------------|
| POST    | /api/users                 | Créer un compte (10 crédits offerts)     |
| GET     | /api/users/{id}            | Profil public d'un utilisateur           |
| PUT     | /api/users/{id}            | Modifier son profil                       |
| GET     | /api/users/{id}/skills     | Compétences d'un utilisateur             |
| PUT     | /api/users/{id}/skills     | Définir ses compétences                   |
| GET     | /api/users/{id}/reviews    | Avis reçus par un utilisateur            |
| GET     | /api/users/{id}/stats      | Statistiques d'un utilisateur            |

### Services

| Méthode | Path                       | Description                              |
|---------|----------------------------|------------------------------------------|
| POST    | /api/services              | Créer une annonce                        |
| GET     | /api/services              | Liste (filtres: `categorie`, `ville`, `search`) |
| GET     | /api/services/{id}         | Détail d'un service                      |
| PUT     | /api/services/{id}         | Modifier son annonce                     |
| DELETE  | /api/services/{id}         | Supprimer son annonce                    |
| GET     | /api/services/{id}/reviews | Avis sur un service                      |

### Échanges

| Méthode | Path                              | Description                        |
|---------|-----------------------------------|------------------------------------|
| POST    | /api/exchanges                    | Créer une demande                  |
| GET     | /api/exchanges                    | Mes échanges (filtre: `status`)    |
| GET     | /api/exchanges/{id}               | Détail d'un échange                |
| PUT     | /api/exchanges/{id}/accept        | Accepter un échange (prestataire)  |
| PUT     | /api/exchanges/{id}/reject        | Refuser un échange (prestataire)   |
| PUT     | /api/exchanges/{id}/complete      | Terminer un échange (demandeur)    |
| PUT     | /api/exchanges/{id}/cancel        | Annuler un échange                 |
| POST    | /api/exchanges/{id}/review        | Laisser un avis                    |

## Exemples curl

```bash
# Créer un utilisateur
curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"pseudo":"Alice","ville":"Paris"}' | jq


# Définir ses compétences puis publier un service
curl -s -X PUT http://localhost:8080/api/users/1/skills \
  -H "Content-Type: application/json" -H "X-User-ID: 1" \
  -d '[{"nom":"Jardinage","niveau":"expert"}]'

curl -s -X POST http://localhost:8080/api/services \
  -H "Content-Type: application/json" -H "X-User-ID: 1" \
  -d '{"titre":"Taille de haies","categorie":"Jardinage","duree_minutes":60,"credits":2}' | jq

# Demander un échange, l'accepter puis le terminer
curl -s -X POST http://localhost:8080/api/exchanges \
  -H "Content-Type: application/json" -H "X-User-ID: 2" \
  -d '{"service_id":1}' | jq

curl -s -X PUT http://localhost:8080/api/exchanges/1/accept -H "X-User-ID: 1" | jq
curl -s -X PUT http://localhost:8080/api/exchanges/1/complete -H "X-User-ID: 2" | jq

# Laisser un avis
curl -s -X POST http://localhost:8080/api/exchanges/1/review \
  -H "Content-Type: application/json" -H "X-User-ID: 2" \
  -d '{"note":5,"commentaire":"Super service !"}' | jq
```

## Tests

```bash
# Lancer les tests avec Docker
docker compose -f compose.test.yml up --build --abort-on-container-exit

# Lancer les tests en local
  go test -v -cover ./...
```
