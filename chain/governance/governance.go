package governance

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

// GovernanceSystem implements on-chain governance for quantum blockchain
type GovernanceSystem struct {
	chainID        *big.Int
	proposals      map[uint64]*Proposal
	votes          map[uint64]map[types.Address]*Vote
	nextProposalID uint64

	// Governance parameters
	votingPeriod    time.Duration
	executionDelay  time.Duration
	quorum          float64  // Minimum participation rate
	threshold       float64  // Minimum approval rate
	proposalDeposit *big.Int // Minimum QTM to create proposal

	// Validator voting power
	validatorSet ValidatorSet

	// Upgrade management
	upgrades        map[string]*NetworkUpgrade
	pendingUpgrades []*NetworkUpgrade

	// Security
	emergencyPause   bool
	emergencyCouncil []types.Address

	// Thread safety
	mu sync.RWMutex

	// Event handlers
	onProposalCreated  func(*Proposal)
	onVoteCast         func(*Vote)
	onProposalPassed   func(*Proposal)
	onUpgradeScheduled func(*NetworkUpgrade)
}

// ValidatorSet interface for voting power calculation
type ValidatorSet interface {
	GetValidator(types.Address) *ValidatorInfo
	GetTotalVotingPower() *big.Int
	GetActiveValidators() []*ValidatorInfo
}

// ValidatorInfo represents validator information for governance
type ValidatorInfo struct {
	Address     types.Address `json:"address"`
	VotingPower *big.Int      `json:"votingPower"`
	IsActive    bool          `json:"isActive"`
}

// Proposal represents a governance proposal
type Proposal struct {
	ID          uint64          `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Proposer    types.Address   `json:"proposer"`
	Type        ProposalType    `json:"type"`
	Content     ProposalContent `json:"content"`

	// Timing
	SubmissionTime time.Time `json:"submissionTime"`
	VotingStart    time.Time `json:"votingStart"`
	VotingEnd      time.Time `json:"votingEnd"`
	ExecutionTime  time.Time `json:"executionTime"`

	// Status
	Status  ProposalStatus `json:"status"`
	Deposit *big.Int       `json:"deposit"`

	// Results
	VotesFor      *big.Int `json:"votesFor"`
	VotesAgainst  *big.Int `json:"votesAgainst"`
	VotesAbstain  *big.Int `json:"votesAbstain"`
	TotalVotes    *big.Int `json:"totalVotes"`
	Participation float64  `json:"participation"`

	// Execution
	Executed        bool   `json:"executed"`
	ExecutionResult string `json:"executionResult,omitempty"`
}

// ProposalType defines different types of proposals
type ProposalType uint8

const (
	ProposalParameterChange ProposalType = iota
	ProposalSoftwareUpgrade
	ProposalValidatorChange
	ProposalEmergencyAction
	ProposalTextProposal
	ProposalTreasurySpend
	ProposalProtocolUpgrade
)

// ProposalStatus represents proposal status
type ProposalStatus uint8

const (
	StatusPending ProposalStatus = iota
	StatusActive
	StatusPassed
	StatusRejected
	StatusExecuted
	StatusFailed
	StatusCancelled
)

// ProposalContent contains proposal-specific data
type ProposalContent struct {
	// Parameter changes
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// Software upgrades
	UpgradeInfo *UpgradeInfo `json:"upgradeInfo,omitempty"`

	// Validator changes
	ValidatorChanges *ValidatorChanges `json:"validatorChanges,omitempty"`

	// Treasury spending
	TreasurySpend *TreasurySpendInfo `json:"treasurySpend,omitempty"`

	// Raw data for custom proposals
	RawData []byte `json:"rawData,omitempty"`
}

// UpgradeInfo defines software upgrade details
type UpgradeInfo struct {
	Name          string                 `json:"name"`
	Version       string                 `json:"version"`
	Height        uint64                 `json:"height"`
	Info          string                 `json:"info"`
	Binaries      map[string]string      `json:"binaries"` // os/arch -> download URL
	Checksum      string                 `json:"checksum"`
	UpgradeParams map[string]interface{} `json:"upgradeParams,omitempty"`
}

// ValidatorChanges defines validator set changes
type ValidatorChanges struct {
	Add    []ValidatorUpdate `json:"add,omitempty"`
	Remove []types.Address   `json:"remove,omitempty"`
	Update []ValidatorUpdate `json:"update,omitempty"`
}

// ValidatorUpdate defines validator updates
type ValidatorUpdate struct {
	Address   types.Address `json:"address"`
	PublicKey []byte        `json:"publicKey"`
	Power     *big.Int      `json:"power"`
}

// TreasurySpendInfo defines treasury spending
type TreasurySpendInfo struct {
	Recipient types.Address `json:"recipient"`
	Amount    *big.Int      `json:"amount"`
	Purpose   string        `json:"purpose"`
}

// Vote represents a governance vote
type Vote struct {
	ProposalID   uint64                    `json:"proposalId"`
	Voter        types.Address             `json:"voter"`
	VotingPower  *big.Int                  `json:"votingPower"`
	Option       VoteOption                `json:"option"`
	Timestamp    time.Time                 `json:"timestamp"`
	Signature    []byte                    `json:"signature"`
	SigAlgorithm crypto.SignatureAlgorithm `json:"sigAlgorithm"`
}

// VoteOption represents voting options
type VoteOption uint8

const (
	VoteYes VoteOption = iota
	VoteNo
	VoteAbstain
	VoteNoWithVeto
)

// NetworkUpgrade represents a scheduled network upgrade
type NetworkUpgrade struct {
	Name      string           `json:"name"`
	Version   string           `json:"version"`
	Height    uint64           `json:"height"`
	Timestamp time.Time        `json:"timestamp"`
	Status    UpgradeStatus    `json:"status"`
	Plan      *UpgradeInfo     `json:"plan"`
	Progress  *UpgradeProgress `json:"progress"`
}

// UpgradeStatus represents upgrade status
type UpgradeStatus uint8

const (
	UpgradeScheduled UpgradeStatus = iota
	UpgradeInProgress
	UpgradeCompleted
	UpgradeFailed
	UpgradeCancelled
)

// UpgradeProgress tracks upgrade progress
type UpgradeProgress struct {
	NodesUpgraded   int       `json:"nodesUpgraded"`
	TotalNodes      int       `json:"totalNodes"`
	ValidatorsReady int       `json:"validatorsReady"`
	TotalValidators int       `json:"totalValidators"`
	LastUpdate      time.Time `json:"lastUpdate"`
}

// NewGovernanceSystem creates a new governance system
func NewGovernanceSystem(chainID *big.Int, validatorSet ValidatorSet) *GovernanceSystem {
	proposalDeposit := new(big.Int)
	proposalDeposit.SetString("10000000000000000000000", 10) // 10,000 QTM

	return &GovernanceSystem{
		chainID:          chainID,
		proposals:        make(map[uint64]*Proposal),
		votes:            make(map[uint64]map[types.Address]*Vote),
		nextProposalID:   1,
		votingPeriod:     7 * 24 * time.Hour, // 7 days
		executionDelay:   24 * time.Hour,     // 1 day delay before execution
		quorum:           0.4,                // 40% participation required
		threshold:        0.5,                // 50% approval required
		proposalDeposit:  proposalDeposit,
		validatorSet:     validatorSet,
		upgrades:         make(map[string]*NetworkUpgrade),
		emergencyCouncil: []types.Address{}, // Set by initialization
	}
}

// SubmitProposal submits a new governance proposal
func (gs *GovernanceSystem) SubmitProposal(
	proposer types.Address,
	title, description string,
	proposalType ProposalType,
	content ProposalContent,
	deposit *big.Int,
) (*Proposal, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Validate deposit
	if deposit.Cmp(gs.proposalDeposit) < 0 {
		return nil, fmt.Errorf("insufficient deposit: required %s QTM", gs.proposalDeposit.String())
	}

	// Validate proposer is active validator
	validator := gs.validatorSet.GetValidator(proposer)
	if validator == nil || !validator.IsActive {
		return nil, errors.New("proposer must be an active validator")
	}

	// Validate proposal content based on type
	if err := gs.validateProposalContent(proposalType, content); err != nil {
		return nil, fmt.Errorf("invalid proposal content: %w", err)
	}

	// Create proposal
	now := time.Now()
	proposal := &Proposal{
		ID:             gs.nextProposalID,
		Title:          title,
		Description:    description,
		Proposer:       proposer,
		Type:           proposalType,
		Content:        content,
		SubmissionTime: now,
		VotingStart:    now.Add(24 * time.Hour), // 1 day review period
		VotingEnd:      now.Add(24*time.Hour + gs.votingPeriod),
		ExecutionTime:  now.Add(24*time.Hour + gs.votingPeriod + gs.executionDelay),
		Status:         StatusPending,
		Deposit:        new(big.Int).Set(deposit),
		VotesFor:       big.NewInt(0),
		VotesAgainst:   big.NewInt(0),
		VotesAbstain:   big.NewInt(0),
		TotalVotes:     big.NewInt(0),
	}

	gs.proposals[proposal.ID] = proposal
	gs.votes[proposal.ID] = make(map[types.Address]*Vote)
	gs.nextProposalID++

	// Trigger event
	if gs.onProposalCreated != nil {
		gs.onProposalCreated(proposal)
	}

	return proposal, nil
}

// CastVote casts a vote on a proposal
func (gs *GovernanceSystem) CastVote(
	proposalID uint64,
	voter types.Address,
	option VoteOption,
	signature []byte,
	sigAlgorithm crypto.SignatureAlgorithm,
) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	proposal, exists := gs.proposals[proposalID]
	if !exists {
		return errors.New("proposal not found")
	}

	// Check voting period
	now := time.Now()
	if now.Before(proposal.VotingStart) {
		return errors.New("voting has not started")
	}
	if now.After(proposal.VotingEnd) {
		return errors.New("voting has ended")
	}

	// Validate voter
	validator := gs.validatorSet.GetValidator(voter)
	if validator == nil || !validator.IsActive {
		return errors.New("voter must be an active validator")
	}

	// Check if already voted
	if _, hasVoted := gs.votes[proposalID][voter]; hasVoted {
		return errors.New("validator has already voted")
	}

	// Verify signature
	voteData := fmt.Sprintf("%d:%s:%d:%d", proposalID, voter.Hex(), option, now.Unix())
	qrSig := &crypto.QRSignature{
		Algorithm: sigAlgorithm,
		Signature: signature,
		PublicKey: validator.Address.Bytes(), // Simplified - would get actual pubkey
	}

	if valid, err := crypto.VerifySignature([]byte(voteData), qrSig); err != nil || !valid {
		return errors.New("invalid vote signature")
	}

	// Create vote
	vote := &Vote{
		ProposalID:   proposalID,
		Voter:        voter,
		VotingPower:  new(big.Int).Set(validator.VotingPower),
		Option:       option,
		Timestamp:    now,
		Signature:    signature,
		SigAlgorithm: sigAlgorithm,
	}

	// Update vote counts
	switch option {
	case VoteYes:
		proposal.VotesFor.Add(proposal.VotesFor, vote.VotingPower)
	case VoteNo, VoteNoWithVeto:
		proposal.VotesAgainst.Add(proposal.VotesAgainst, vote.VotingPower)
	case VoteAbstain:
		proposal.VotesAbstain.Add(proposal.VotesAbstain, vote.VotingPower)
	}

	proposal.TotalVotes.Add(proposal.TotalVotes, vote.VotingPower)

	// Update participation rate
	totalVotingPower := gs.validatorSet.GetTotalVotingPower()
	if totalVotingPower.Sign() > 0 {
		participation := new(big.Int).Set(proposal.TotalVotes)
		participation.Mul(participation, big.NewInt(10000))
		participation.Div(participation, totalVotingPower)
		proposal.Participation = float64(participation.Uint64()) / 100.0
	}

	// Store vote
	gs.votes[proposalID][voter] = vote

	// Trigger event
	if gs.onVoteCast != nil {
		gs.onVoteCast(vote)
	}

	return nil
}

// TallyVotes tallies votes for a proposal and updates its status
func (gs *GovernanceSystem) TallyVotes(proposalID uint64) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	proposal, exists := gs.proposals[proposalID]
	if !exists {
		return errors.New("proposal not found")
	}

	if proposal.Status != StatusActive && proposal.Status != StatusPending {
		return errors.New("proposal is not in voting phase")
	}

	now := time.Now()
	if now.Before(proposal.VotingEnd) {
		return errors.New("voting period has not ended")
	}

	// Check quorum
	if proposal.Participation < gs.quorum {
		proposal.Status = StatusRejected
		return nil
	}

	// Check approval threshold
	totalDecisionVotes := new(big.Int).Add(proposal.VotesFor, proposal.VotesAgainst)
	if totalDecisionVotes.Sign() == 0 {
		proposal.Status = StatusRejected
		return nil
	}

	// Calculate approval rate
	approvalRate := new(big.Int).Set(proposal.VotesFor)
	approvalRate.Mul(approvalRate, big.NewInt(10000))
	approvalRate.Div(approvalRate, totalDecisionVotes)
	approvalRateFloat := float64(approvalRate.Uint64()) / 10000.0

	// Determine result
	if approvalRateFloat >= gs.threshold {
		proposal.Status = StatusPassed

		// Trigger event
		if gs.onProposalPassed != nil {
			gs.onProposalPassed(proposal)
		}

		// Schedule execution for upgrade proposals
		if proposal.Type == ProposalSoftwareUpgrade && proposal.Content.UpgradeInfo != nil {
			err := gs.scheduleUpgrade(proposal.Content.UpgradeInfo)
			if err != nil {
				return fmt.Errorf("failed to schedule upgrade: %w", err)
			}
		}
	} else {
		proposal.Status = StatusRejected
	}

	return nil
}

// ExecuteProposal executes a passed proposal
func (gs *GovernanceSystem) ExecuteProposal(proposalID uint64) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	proposal, exists := gs.proposals[proposalID]
	if !exists {
		return errors.New("proposal not found")
	}

	if proposal.Status != StatusPassed {
		return errors.New("proposal has not passed")
	}

	if proposal.Executed {
		return errors.New("proposal already executed")
	}

	now := time.Now()
	if now.Before(proposal.ExecutionTime) {
		return errors.New("execution delay has not passed")
	}

	// Execute based on proposal type
	var err error
	switch proposal.Type {
	case ProposalParameterChange:
		err = gs.executeParameterChange(proposal)
	case ProposalSoftwareUpgrade:
		err = gs.executeSoftwareUpgrade(proposal)
	case ProposalValidatorChange:
		err = gs.executeValidatorChange(proposal)
	case ProposalTreasurySpend:
		err = gs.executeTreasurySpend(proposal)
	default:
		err = errors.New("unsupported proposal type for automatic execution")
	}

	if err != nil {
		proposal.Status = StatusFailed
		proposal.ExecutionResult = err.Error()
		return err
	}

	proposal.Executed = true
	proposal.Status = StatusExecuted
	proposal.ExecutionResult = "Successfully executed"

	return nil
}

// scheduleUpgrade schedules a network upgrade
func (gs *GovernanceSystem) scheduleUpgrade(upgradeInfo *UpgradeInfo) error {
	upgrade := &NetworkUpgrade{
		Name:      upgradeInfo.Name,
		Version:   upgradeInfo.Version,
		Height:    upgradeInfo.Height,
		Timestamp: time.Now(),
		Status:    UpgradeScheduled,
		Plan:      upgradeInfo,
		Progress: &UpgradeProgress{
			LastUpdate: time.Now(),
		},
	}

	gs.upgrades[upgrade.Name] = upgrade
	gs.pendingUpgrades = append(gs.pendingUpgrades, upgrade)

	// Sort by height
	sort.Slice(gs.pendingUpgrades, func(i, j int) bool {
		return gs.pendingUpgrades[i].Height < gs.pendingUpgrades[j].Height
	})

	// Trigger event
	if gs.onUpgradeScheduled != nil {
		gs.onUpgradeScheduled(upgrade)
	}

	return nil
}

// validateProposalContent validates proposal content based on type
func (gs *GovernanceSystem) validateProposalContent(proposalType ProposalType, content ProposalContent) error {
	switch proposalType {
	case ProposalParameterChange:
		if len(content.Parameters) == 0 {
			return errors.New("parameter change proposals must specify parameters")
		}
	case ProposalSoftwareUpgrade:
		if content.UpgradeInfo == nil {
			return errors.New("software upgrade proposals must include upgrade info")
		}
		if content.UpgradeInfo.Name == "" || content.UpgradeInfo.Version == "" {
			return errors.New("upgrade info must include name and version")
		}
	case ProposalValidatorChange:
		if content.ValidatorChanges == nil {
			return errors.New("validator change proposals must include validator changes")
		}
	case ProposalTreasurySpend:
		if content.TreasurySpend == nil {
			return errors.New("treasury spend proposals must include spend info")
		}
		if content.TreasurySpend.Amount.Sign() <= 0 {
			return errors.New("treasury spend amount must be positive")
		}
	}
	return nil
}

// Execution methods (simplified implementations)
func (gs *GovernanceSystem) executeParameterChange(proposal *Proposal) error {
	// Implementation would update chain parameters
	return nil
}

func (gs *GovernanceSystem) executeSoftwareUpgrade(proposal *Proposal) error {
	// Implementation would trigger upgrade process
	return nil
}

func (gs *GovernanceSystem) executeValidatorChange(proposal *Proposal) error {
	// Implementation would update validator set
	return nil
}

func (gs *GovernanceSystem) executeTreasurySpend(proposal *Proposal) error {
	// Implementation would transfer treasury funds
	return nil
}

// GetProposal returns a proposal by ID
func (gs *GovernanceSystem) GetProposal(proposalID uint64) (*Proposal, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	proposal, exists := gs.proposals[proposalID]
	if !exists {
		return nil, errors.New("proposal not found")
	}

	return proposal, nil
}

// GetProposals returns all proposals with optional filtering
func (gs *GovernanceSystem) GetProposals(status *ProposalStatus, proposalType *ProposalType) []*Proposal {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	proposals := make([]*Proposal, 0)
	for _, proposal := range gs.proposals {
		if status != nil && proposal.Status != *status {
			continue
		}
		if proposalType != nil && proposal.Type != *proposalType {
			continue
		}
		proposals = append(proposals, proposal)
	}

	// Sort by ID (newest first)
	sort.Slice(proposals, func(i, j int) bool {
		return proposals[i].ID > proposals[j].ID
	})

	return proposals
}

// GetUpgrades returns all network upgrades
func (gs *GovernanceSystem) GetUpgrades() []*NetworkUpgrade {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	upgrades := make([]*NetworkUpgrade, 0, len(gs.upgrades))
	for _, upgrade := range gs.upgrades {
		upgrades = append(upgrades, upgrade)
	}

	return upgrades
}

// GetPendingUpgrades returns pending upgrades
func (gs *GovernanceSystem) GetPendingUpgrades() []*NetworkUpgrade {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	result := make([]*NetworkUpgrade, len(gs.pendingUpgrades))
	copy(result, gs.pendingUpgrades)
	return result
}

// GetGovernanceParams returns current governance parameters
func (gs *GovernanceSystem) GetGovernanceParams() map[string]interface{} {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return map[string]interface{}{
		"votingPeriod":    gs.votingPeriod.Hours(),
		"executionDelay":  gs.executionDelay.Hours(),
		"quorum":          gs.quorum,
		"threshold":       gs.threshold,
		"proposalDeposit": gs.proposalDeposit.String(),
		"emergencyPause":  gs.emergencyPause,
		"totalProposals":  gs.nextProposalID - 1,
		"activeProposals": gs.countActiveProposals(),
	}
}

// countActiveProposals counts active proposals
func (gs *GovernanceSystem) countActiveProposals() int {
	count := 0
	now := time.Now()
	for _, proposal := range gs.proposals {
		if now.After(proposal.VotingStart) && now.Before(proposal.VotingEnd) {
			count++
		}
	}
	return count
}

// SetEventHandlers sets governance event handlers
func (gs *GovernanceSystem) SetEventHandlers(
	onProposalCreated func(*Proposal),
	onVoteCast func(*Vote),
	onProposalPassed func(*Proposal),
	onUpgradeScheduled func(*NetworkUpgrade),
) {
	gs.onProposalCreated = onProposalCreated
	gs.onVoteCast = onVoteCast
	gs.onProposalPassed = onProposalPassed
	gs.onUpgradeScheduled = onUpgradeScheduled
}
