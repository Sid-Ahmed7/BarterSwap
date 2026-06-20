package main

// User represents a platform user.
type User struct {
	ID            int     `json:"id"`
	Pseudo        string  `json:"pseudo"`
	Bio           string  `json:"bio,omitempty"`
	Ville         string  `json:"ville,omitempty"`
	Skills        []Skill `json:"skills,omitempty"`
	CreditBalance int     `json:"credit_balance"`
	CreatedAt     string  `json:"created_at"`
}

// Skill represents a user competency with a name and level.
type Skill struct {
	Nom    string `json:"nom"`
	Niveau string `json:"niveau"`
}

// UserRequest holds the fields for creating or updating a user.
type UserRequest struct {
	Pseudo string `json:"pseudo"`
	Bio    string `json:"bio"`
	Ville  string `json:"ville"`
}

// Service represents a skill-sharing listing posted by a user.
type Service struct {
	ID           int    `json:"id"`
	ProviderID   int    `json:"provider_id"`
	Titre        string `json:"titre"`
	Description  string `json:"description,omitempty"`
	Categorie    string `json:"categorie"`
	DureeMinutes int    `json:"duree_minutes"`
	Credits      int    `json:"credits"`
	Ville        string `json:"ville,omitempty"`
	Actif        bool   `json:"actif"`
	CreatedAt    string `json:"created_at"`
}

// ServiceRequest holds the fields for creating or updating a service.
type ServiceRequest struct {
	Titre        string `json:"titre"`
	Description  string `json:"description"`
	Categorie    string `json:"categorie"`
	DureeMinutes int    `json:"duree_minutes"`
	Credits      int    `json:"credits"`
	Ville        string `json:"ville"`
}

// ServiceListRequest holds the optional filters for listing services.
type ServiceListRequest struct {
	Categorie string `json:"categorie"`
	Ville     string `json:"ville"`
	Search    string `json:"search"`
}

// Exchange represents a service exchange request between two users.
type Exchange struct {
	ID          int    `json:"id"`
	ServiceID   int    `json:"service_id"`
	RequesterID int    `json:"requester_id"`
	OwnerID     int    `json:"owner_id"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CreditTransaction records every credit movement for an exchange.
type CreditTransaction struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	ExchangeID int    `json:"exchange_id"`
	Montant    int    `json:"montant"`
	Type       string `json:"type"`
	CreatedAt  string `json:"created_at"`
}

// ExchangeRequest holds the fields for creating an exchange.
type ExchangeRequest struct {
	ServiceID   int `json:"service_id"`
	RequesterID int `json:"requester_id"`
	OwnerID     int `json:"owner_id"`
}

// Review represents a rating left by one party after a completed exchange.
type Review struct {
	ID          int    `json:"id"`
	ExchangeID  int    `json:"exchange_id"`
	AuthorID    int    `json:"author_id"`
	TargetID    int    `json:"target_id"`
	Note        int    `json:"note"`
	Commentaire string `json:"commentaire,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// ReviewRequest holds the fields for submitting a review (note 1-5).
type ReviewRequest struct {
	Note        int    `json:"note"`
	Commentaire string `json:"commentaire"`
}

// UserStats aggregates activity statistics for a user.
type UserStats struct {
	UserID            int     `json:"user_id"`
	ServicesActifs    int     `json:"services_actifs"`
	EchangesCompletes int     `json:"echanges_completes"`
	CreditBalance     int     `json:"credit_balance"`
	NoteMoyenne       float64 `json:"note_moyenne"`
	NbAvis            int     `json:"nb_avis"`
	TotalGagne        int     `json:"total_gagne"`
	TotalDepense      int     `json:"total_depense"`
}
