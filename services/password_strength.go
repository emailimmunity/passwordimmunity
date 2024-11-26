package services

import (
	"math"
	"strings"
	"unicode"
)

type PasswordStrengthService interface {
	EvaluateStrength(password string) *PasswordStrength
	GetStrengthRequirements(orgID uuid.UUID) (*StrengthRequirements, error)
	UpdateStrengthRequirements(orgID uuid.UUID, requirements *StrengthRequirements) error
}

type PasswordStrength struct {
	Score           float64  // 0-100
	Entropy         float64  // Bits of entropy
	CrackTime      string   // Estimated time to crack
	Weaknesses     []string // List of identified weaknesses
	Recommendations []string // Suggestions for improvement
}

type StrengthRequirements struct {
	MinimumScore    float64
	MinimumEntropy  float64
	RequiredPatterns []string
}

type passwordStrengthService struct {
	repo     repository.Repository
	cache    CacheService
	patterns map[string]*regexp.Regexp
}

func NewPasswordStrengthService(
	repo repository.Repository,
	cache CacheService,
) PasswordStrengthService {
	return &passwordStrengthService{
		repo:  repo,
		cache: cache,
		patterns: initializePatterns(),
	}
}

func (s *passwordStrengthService) EvaluateStrength(password string) *PasswordStrength {
	strength := &PasswordStrength{}

	// Calculate base entropy
	entropy := s.calculateEntropy(password)
	strength.Entropy = entropy

	// Calculate score based on multiple factors
	score := s.calculateScore(password, entropy)
	strength.Score = score

	// Estimate crack time
	strength.CrackTime = s.estimateCrackTime(entropy)

	// Identify weaknesses
	strength.Weaknesses = s.identifyWeaknesses(password)

	// Generate recommendations
	strength.Recommendations = s.generateRecommendations(password, strength.Weaknesses)

	return strength
}

func (s *passwordStrengthService) calculateEntropy(password string) float64 {
	// Character set size calculation
	var poolSize float64

	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasDigit := strings.ContainsAny(password, "0123456789")
	hasSpecial := false
	for _, r := range password {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			hasSpecial = true
			break
		}
	}

	if hasLower {
		poolSize += 26
	}
	if hasUpper {
		poolSize += 26
	}
	if hasDigit {
		poolSize += 10
	}
	if hasSpecial {
		poolSize += 32
	}

	return float64(len(password)) * math.Log2(poolSize)
}

func (s *passwordStrengthService) calculateScore(password string, entropy float64) float64 {
	score := 0.0

	// Base score from entropy
	score += entropy * 2

	// Length bonus
	score += float64(len(password)) * 4

	// Character variety bonus
	if strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		score += 10
	}
	if strings.ContainsAny(password, "0123456789") {
		score += 10
	}
	if strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?") {
		score += 15
	}

	// Pattern penalties
	if s.containsCommonPatterns(password) {
		score -= 20
	}

	// Normalize score to 0-100 range
	score = math.Max(0, math.Min(100, score))

	return score
}

func (s *passwordStrengthService) estimateCrackTime(entropy float64) string {
	// Assume 10^12 guesses per second (modern hardware)
	guesses := math.Pow(2, entropy)
	seconds := guesses / 1e12

	switch {
	case seconds < 1:
		return "instant"
	case seconds < 60:
		return "seconds"
	case seconds < 3600:
		return "minutes"
	case seconds < 86400:
		return "hours"
	case seconds < 2592000:
		return "days"
	case seconds < 31536000:
		return "months"
	default:
		years := seconds / 31536000
		if years > 1000000 {
			return "centuries"
		}
		return fmt.Sprintf("%.0f years", years)
	}
}

func (s *passwordStrengthService) identifyWeaknesses(password string) []string {
	var weaknesses []string

	if len(password) < 12 {
		weaknesses = append(weaknesses, "Password is too short")
	}

	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		weaknesses = append(weaknesses, "Missing uppercase letters")
	}

	if !strings.ContainsAny(password, "0123456789") {
		weaknesses = append(weaknesses, "Missing numbers")
	}


	if !strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?") {
		weaknesses = append(weaknesses, "Missing special characters")
	}

	if s.containsCommonPatterns(password) {
		weaknesses = append(weaknesses, "Contains common patterns")
	}

	return weaknesses
}

func (s *passwordStrengthService) generateRecommendations(password string, weaknesses []string) []string {
	var recommendations []string

	for _, weakness := range weaknesses {
		switch weakness {
		case "Password is too short":
			recommendations = append(recommendations, "Increase password length to at least 12 characters")
		case "Missing uppercase letters":
			recommendations = append(recommendations, "Add uppercase letters")
		case "Missing numbers":
			recommendations = append(recommendations, "Add numbers")
		case "Missing special characters":
			recommendations = append(recommendations, "Add special characters")
		case "Contains common patterns":
			recommendations = append(recommendations, "Avoid common patterns and sequences")
		}
	}

	return recommendations
}

func (s *passwordStrengthService) GetStrengthRequirements(orgID uuid.UUID) (*StrengthRequirements, error) {
	cacheKey := fmt.Sprintf("strength_requirements:%s", orgID)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != nil {
		return cached.(*StrengthRequirements), nil
	}

	requirements, err := s.repo.GetStrengthRequirements(orgID)
	if err != nil {
		return nil, err
	}

	if requirements != nil {
		s.cache.Set(ctx, cacheKey, requirements, time.Hour*24)
	}

	return requirements, nil
}

func (s *passwordStrengthService) UpdateStrengthRequirements(orgID uuid.UUID, requirements *StrengthRequirements) error {
	if err := s.repo.UpdateStrengthRequirements(orgID, requirements); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("strength_requirements:%s", orgID)
	s.cache.Delete(ctx, cacheKey)

	return nil
}

func initializePatterns() map[string]*regexp.Regexp {
	return map[string]*regexp.Regexp{
		"keyboard_sequence": regexp.MustCompile(`(?i)(qwerty|asdfgh|zxcvbn)`),
		"number_sequence":   regexp.MustCompile(`(?:0123|1234|2345|3456|4567|5678|6789|7890)`),
		"letter_sequence":   regexp.MustCompile(`(?i)(?:abcd|bcde|cdef|defg|efgh|fghi|ghij|hijk|ijkl|jklm|klmn|lmno|mnop|nopq|opqr|pqrs|qrst|rstu|stuv|tuvw|uvwx|vwxy|wxyz)`),
		"repeated_chars":    regexp.MustCompile(`(.)\1{2,}`),
	}
}

func (s *passwordStrengthService) containsCommonPatterns(password string) bool {
	for _, pattern := range s.patterns {
		if pattern.MatchString(password) {
			return true
		}
	}
	return false
}
