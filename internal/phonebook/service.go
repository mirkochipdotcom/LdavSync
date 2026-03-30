package phonebook

import (
	"github.com/mirkochipdotcom/gorubrica/internal/database"
)

// ContactWithGroups represents a contact with its associated groups
type ContactWithGroups struct {
	Contact *database.Contact
	Groups  []*database.GroupNumber
}

// Service provides business logic for phonebook operations
type Service struct {
	db *database.DB
}

// NewService creates a new phonebook service
func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// GetContactWithGroups retrieves a contact with all its group memberships
func (s *Service) GetContactWithGroups(uid string) (*ContactWithGroups, error) {
	contact, err := s.db.GetContact(uid)
	if err != nil {
		return nil, err
	}
	if contact == nil {
		return nil, nil
	}

	groups, err := s.db.GetContactGroups(contact.ID)
	if err != nil {
		return nil, err
	}

	return &ContactWithGroups{
		Contact: contact,
		Groups:  groups,
	}, nil
}

// SearchContactsWithGroups searches contacts and includes their groups
func (s *Service) SearchContactsWithGroups(query string, limit int) ([]*ContactWithGroups, error) {
	contacts, err := s.db.SearchContacts(query, limit)
	if err != nil {
		return nil, err
	}

	results := make([]*ContactWithGroups, 0, len(contacts))
	for _, contact := range contacts {
		groups, err := s.db.GetContactGroups(contact.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, &ContactWithGroups{
			Contact: contact,
			Groups:  groups,
		})
	}

	return results, nil
}

// ListContactsWithGroups lists contacts with pagination and includes their groups
func (s *Service) ListContactsWithGroups(limit, offset int) ([]*ContactWithGroups, error) {
	contacts, err := s.db.ListContacts(limit, offset)
	if err != nil {
		return nil, err
	}

	results := make([]*ContactWithGroups, 0, len(contacts))
	for _, contact := range contacts {
		groups, err := s.db.GetContactGroups(contact.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, &ContactWithGroups{
			Contact: contact,
			Groups:  groups,
		})
	}

	return results, nil
}

// GroupWithMembers represents a group with its members
type GroupWithMembers struct {
	Group   *database.GroupNumber
	Members []*database.Contact
}

// GetGroupWithMembers retrieves a group with all its members
func (s *Service) GetGroupWithMembers(id int64) (*GroupWithMembers, error) {
	group, err := s.db.GetGroup(id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}

	members, err := s.db.GetGroupMembers(id)
	if err != nil {
		return nil, err
	}

	return &GroupWithMembers{
		Group:   group,
		Members: members,
	}, nil
}

// ListGroupsWithMembers lists all groups with their member counts
func (s *Service) ListGroupsWithMembers() ([]*GroupWithMembers, error) {
	groups, err := s.db.ListGroups()
	if err != nil {
		return nil, err
	}

	results := make([]*GroupWithMembers, 0, len(groups))
	for _, group := range groups {
		members, err := s.db.GetGroupMembers(group.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, &GroupWithMembers{
			Group:   group,
			Members: members,
		})
	}

	return results, nil
}
