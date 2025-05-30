-- Users Table
CREATE TABLE users (
  user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(255) NOT NULL UNIQUE,
  email VARCHAR(255) UNIQUE,
  display_name VARCHAR(255),
  avatar_url TEXT,
  avatar_source VARCHAR(50) DEFAULT NULL,
  stats_privacy VARCHAR(50) NOT NULL DEFAULT 'public' CHECK (stats_privacy IN ('private', 'friends_only', 'public')),
  ui_theme VARCHAR(50) DEFAULT 'system',
  color_theme VARCHAR(50) DEFAULT 'blue',
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  last_login_at TIMESTAMPTZ,
  CONSTRAINT uq_users_username UNIQUE (username),
  CONSTRAINT uq_users_email UNIQUE (email)
);

-- Provider Identities Table
-- Stores information about each OAuth identity linked to a user
CREATE TABLE user_provider_identities (
  provider_identity_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  provider_name VARCHAR(50) NOT NULL, -- e.g., 'google', 'twitter', 'instagram'
  provider_user_id TEXT NOT NULL,     -- The unique ID for the user *on the provider's system*
  provider_email VARCHAR(255),        -- Email from provider, might not be user's primary email
  provider_display_name VARCHAR(255),
  provider_avatar_url TEXT,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uq_user_providers_identities_provider UNIQUE (provider_name, provider_user_id), -- A user can only link a specific provider account once
  CONSTRAINT uq_user_providers_identities_user_provider UNIQUE (user_id, provider_name)          
);

-- User Friendships Table
-- Stores the relationship between two users
CREATE TABLE user_friendships (
  friendship_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  requester_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  addressee_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'blocked')),
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT chk_different_users_friendship CHECK (requester_id <> addressee_id),
  CONSTRAINT uq_requester_addressee UNIQUE (requester_id, addressee_id)
);

-- Game Sessions Table
CREATE TABLE game_sessions (
  session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_name VARCHAR(255),
  created_by_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'abandoned')),
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  completed_at TIMESTAMPTZ
);

-- Guest Players Table
CREATE TABLE guest_players (
  guest_player_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  display_name VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Define game_players table before games table because games will reference it
-- The foreign key from game_players to games wil lbe added after via ALTER TABLE
CREATE TABLE game_players (
  game_player_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  game_id UUID NOT NULL,
  user_id UUID REFERENCES users(user_id) ON DELETE CASCADE, -- If a user is deleted, their game player entries are removed
  guest_player_id UUID REFERENCES guest_players(guest_player_id) ON DELETE CASCADE,
  seating_order INTEGER NOT NULL CHECK (seating_order > 0),
  final_score INTEGER NOT NULL DEFAULT 0,
  finishing_position INTEGER CHECK (finishing_position IS NULL OR finishing_position > 0),
  CONSTRAINT uq_game_user UNIQUE (game_id, user_id),
  CONSTRAINT uq_game_guest UNIQUE (game_id, guest_player_id),
  CONSTRAINT chk_player_type CHECK (user_id IS NOT NULL OR guest_player_id IS NOT NULL)
);

-- Games Table
CREATE TABLE games (
  game_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID REFERENCES game_sessions(session_id) ON DELETE SET NULL,
  created_by_user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
  current_scorekeeper_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'active', 'completed', 'abandoned')),
  starting_dealer_game_player_id UUID REFERENCES game_players(game_player_id) ON DELETE SET NULL,
  player_seating_order_randomized BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  completed_at TIMESTAMPTZ
);

-- Add the foreign key from game_players to games now that the game table exists
ALTER TABLE game_players
ADD CONSTRAINT fk_game_players_game_id
FOREIGN KEY (game_id)
REFERENCES games(game_id)
ON DELETE CASCADE; -- If a game is deleted, all its player associations are deleted

-- Rounds Table
CREATE TABLE rounds (
  round_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  game_id UUID NOT NULL REFERENCES games(game_id) ON DELETE CASCADE,
  round_number INTEGER NOT NULL CHECK (round_number > 0),
  dealer_game_player_id UUID NOT NULL REFERENCES game_players(game_player_id) ON DELETE RESTRICT,
  status VARCHAR(50) NOT NULL CHECK (status IN ('bidding', 'playing', 'completed')),
  is_tiebreaker_round BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uq_game_round UNIQUE (game_id, round_number)
);

-- Player Round Scores Table
CREATE TABLE player_round_scores (
  player_round_score_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  round_id UUID NOT NULL REFERENCES rounds(round_id) ON DELETE CASCADE,
  game_player_id UUID NOT NULL REFERENCES game_players(game_player_id) ON DELETE CASCADE,
  bid_amount INTEGER NOT NULL CHECK (bid_amount >= 0),
  tricks_taken INTEGER CHECK (tricks_taken IS NULL OR tricks_taken >= 0),
  round_score INTEGER NOT NULL DEFAULT 0,
  bonus_points_applied INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uq_round_player UNIQUE (round_id, game_player_id)
);

-- Player Game Asterisks Table
CREATE TABLE player_game_asterisks (
  player_game_asterisk_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  game_player_id UUID NOT NULL REFERENCES game_players(game_player_id) ON DELETE CASCADE,
  game_id UUID NOT NULL REFERENCES games(game_id) ON DELETE CASCADE,
  reason VARCHAR(255),
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- User Notifications Table
CREATE TABLE user_notifications (
  notification_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  recipient_user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  type VARCHAR(50) NOT NULL,
  actor_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  message TEXT NOT NULL,
  is_read BOOLEAN NOT NULL DEFAULT FALSE,
  link TEXT,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Functions to update 'updated_at' timestamps
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to tables with 'updated_at'
CREATE TRIGGER set_timestamp_users
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp_user_provider_identities
BEFORE UPDATE ON user_provider_identities
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp_user_friendships
BEFORE UPDATE ON user_friendships
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp_game_sessions
BEFORE UPDATE ON game_sessions
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp_games
BEFORE UPDATE ON games
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp_rounds
BEFORE UPDATE ON rounds
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp_player_round_scores
BEFORE UPDATE ON player_round_scores
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- Indexes
CREATE INDEX idx_user_provider_identities_user_id ON user_provider_identities(user_id);
CREATE INDEX idx_user_provider_identities_provider_lookup ON user_provider_identities(provider_name, provider_user_id);

CREATE INDEX idx_user_friendships_requester_id ON user_friendships(requester_id);
CREATE INDEX idx_user_friendships_addressee_id ON user_friendships(addressee_id);
CREATE INDEX idx_user_friendships_status ON user_friendships(status);

CREATE INDEX idx_game_sessions_created_by_user_id ON game_sessions(created_by_user_id);

CREATE INDEX idx_games_session_id ON games(session_id);
CREATE INDEX idx_games_created_by_user_id ON games(created_by_user_id);
CREATE INDEX idx_games_starting_dealer_game_player_id ON games(starting_dealer_game_player_id);

CREATE INDEX idx_game_players_game_id ON game_players(game_id);
CREATE INDEX idx_game_players_user_id ON game_players(user_id);
CREATE INDEX idx_game_players_guest_player_id ON game_players(guest_player_id);

CREATE INDEX idx_rounds_game_id ON rounds(game_id);
CREATE INDEX idx_rounds_dealer_game_player_id ON rounds(dealer_game_player_id);

CREATE INDEX idx_player_round_scores_round_id ON player_round_scores(round_id);
CREATE INDEX idx_player_round_scores_game_player_id ON player_round_scores(game_player_id);

CREATE INDEX idx_player_game_asterisks_game_player_id ON player_game_asterisks(game_player_id);
CREATE INDEX idx_player_game_asterisks_game_id ON player_game_asterisks(game_id);

CREATE INDEX idx_user_notifications_recipient_user_id ON user_notifications(recipient_user_id);
CREATE INDEX idx_user_notifications_is_read ON user_notifications(recipient_user_id, is_read);
CREATE INDEX idx_user_notifications_actor_user_id ON user_notifications(actor_user_id);
