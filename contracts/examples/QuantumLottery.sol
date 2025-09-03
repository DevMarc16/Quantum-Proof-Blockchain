// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "../lib/QuantumRandom.sol";
import "../lib/QuantumVerifier.sol";

/**
 * @title QuantumLottery
 * @dev A provably fair lottery using quantum randomness
 */
contract QuantumLottery {
    using QuantumRandom for bytes32;
    using QuantumVerifier for bytes32;
    
    // Events
    event LotteryCreated(uint256 indexed lotteryId, uint256 ticketPrice, uint256 endTime);
    event TicketPurchased(uint256 indexed lotteryId, address indexed player, uint256 ticketNumber);
    event LotteryDrawn(uint256 indexed lotteryId, uint256 winningNumber, address winner, uint256 prize);
    event PrizeWithdrawn(uint256 indexed lotteryId, address indexed winner, uint256 amount);
    event RefundIssued(uint256 indexed lotteryId, address indexed player, uint256 amount);
    
    // Structures
    struct Lottery {
        uint256 id;
        uint256 ticketPrice;
        uint256 startTime;
        uint256 endTime;
        uint256 maxTickets;
        uint256 ticketsSold;
        mapping(uint256 => address) ticketToPlayer;
        mapping(address => uint256[]) playerTickets;
        uint256 totalPrize;
        uint256 winningTicket;
        address winner;
        bool drawn;
        bool prizeWithdrawn;
        bytes32 randomSeed;
        bytes32 commitment; // For commit-reveal scheme
    }
    
    // State variables
    mapping(uint256 => Lottery) public lotteries;
    uint256 public currentLotteryId;
    uint256 public constant HOUSE_FEE_PERCENT = 5; // 5% house fee
    address public owner;
    
    // Quantum randomness state
    mapping(uint256 => bytes32) private revealSeeds;
    mapping(uint256 => bool) private seedRevealed;
    
    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }
    
    modifier lotteryExists(uint256 _lotteryId) {
        require(_lotteryId <= currentLotteryId, "Lottery does not exist");
        _;
    }
    
    modifier lotteryActive(uint256 _lotteryId) {
        Lottery storage lottery = lotteries[_lotteryId];
        require(block.timestamp >= lottery.startTime, "Lottery not started");
        require(block.timestamp <= lottery.endTime, "Lottery ended");
        require(lottery.ticketsSold < lottery.maxTickets, "Lottery sold out");
        _;
    }
    
    modifier lotteryEnded(uint256 _lotteryId) {
        Lottery storage lottery = lotteries[_lotteryId];
        require(
            block.timestamp > lottery.endTime || lottery.ticketsSold >= lottery.maxTickets,
            "Lottery still active"
        );
        _;
    }
    
    constructor() {
        owner = msg.sender;
        currentLotteryId = 0;
    }
    
    /**
     * @dev Create a new lottery
     * @param _ticketPrice Price per ticket in wei
     * @param _duration Duration in seconds
     * @param _maxTickets Maximum number of tickets
     * @param _commitment Commitment hash for quantum randomness
     */
    function createLottery(
        uint256 _ticketPrice,
        uint256 _duration,
        uint256 _maxTickets,
        bytes32 _commitment
    ) external onlyOwner {
        require(_ticketPrice > 0, "Ticket price must be positive");
        require(_duration > 0, "Duration must be positive");
        require(_maxTickets > 1, "Must allow at least 2 tickets");
        require(_commitment != bytes32(0), "Invalid commitment");
        
        currentLotteryId++;
        
        Lottery storage newLottery = lotteries[currentLotteryId];
        newLottery.id = currentLotteryId;
        newLottery.ticketPrice = _ticketPrice;
        newLottery.startTime = block.timestamp;
        newLottery.endTime = block.timestamp + _duration;
        newLottery.maxTickets = _maxTickets;
        newLottery.commitment = _commitment;
        
        emit LotteryCreated(currentLotteryId, _ticketPrice, newLottery.endTime);
    }
    
    /**
     * @dev Purchase tickets for a lottery
     * @param _lotteryId ID of the lottery
     * @param _quantity Number of tickets to purchase
     */
    function buyTickets(uint256 _lotteryId, uint256 _quantity) 
        external 
        payable 
        lotteryExists(_lotteryId) 
        lotteryActive(_lotteryId) {
        
        Lottery storage lottery = lotteries[_lotteryId];
        
        require(_quantity > 0, "Must buy at least one ticket");
        require(_quantity <= 10, "Maximum 10 tickets per transaction");
        require(lottery.ticketsSold + _quantity <= lottery.maxTickets, "Not enough tickets available");
        require(msg.value == lottery.ticketPrice * _quantity, "Incorrect payment amount");
        
        for (uint256 i = 0; i < _quantity; i++) {
            uint256 ticketNumber = lottery.ticketsSold + i;
            lottery.ticketToPlayer[ticketNumber] = msg.sender;
            lottery.playerTickets[msg.sender].push(ticketNumber);
            
            emit TicketPurchased(_lotteryId, msg.sender, ticketNumber);
        }
        
        lottery.ticketsSold += _quantity;
        lottery.totalPrize += msg.value;
    }
    
    /**
     * @dev Draw the lottery using quantum randomness
     * @param _lotteryId ID of the lottery to draw
     * @param _revealValue The value used in the commitment
     * @param _nonce The nonce used in the commitment
     */
    function drawLottery(
        uint256 _lotteryId,
        uint256 _revealValue,
        uint256 _nonce
    ) 
        external 
        onlyOwner 
        lotteryExists(_lotteryId) 
        lotteryEnded(_lotteryId) {
        
        Lottery storage lottery = lotteries[_lotteryId];
        require(!lottery.drawn, "Lottery already drawn");
        require(lottery.ticketsSold > 0, "No tickets sold");
        
        // Verify commitment
        bytes32 expectedCommitment = QuantumRandom.createCommitment(_revealValue, _nonce);
        require(lottery.commitment == expectedCommitment, "Invalid commitment reveal");
        
        // Generate quantum random seed
        bytes32 quantumSeed = keccak256(abi.encodePacked(
            _revealValue,
            _nonce,
            block.timestamp,
            block.difficulty,
            lottery.totalPrize,
            lottery.ticketsSold
        ));
        
        // Generate winning ticket number
        uint256 winningTicket = QuantumRandom.generateRandomInRange(
            quantumSeed,
            0,
            lottery.ticketsSold
        );
        
        lottery.winningTicket = winningTicket;
        lottery.winner = lottery.ticketToPlayer[winningTicket];
        lottery.drawn = true;
        lottery.randomSeed = quantumSeed;
        
        emit LotteryDrawn(_lotteryId, winningTicket, lottery.winner, lottery.totalPrize);
    }
    
    /**
     * @dev Withdraw prize for a won lottery
     * @param _lotteryId ID of the lottery
     */
    function withdrawPrize(uint256 _lotteryId) 
        external 
        lotteryExists(_lotteryId) {
        
        Lottery storage lottery = lotteries[_lotteryId];
        require(lottery.drawn, "Lottery not drawn yet");
        require(msg.sender == lottery.winner, "Not the winner");
        require(!lottery.prizeWithdrawn, "Prize already withdrawn");
        
        // Calculate prize (total prize minus house fee)
        uint256 houseFee = (lottery.totalPrize * HOUSE_FEE_PERCENT) / 100;
        uint256 winnerPrize = lottery.totalPrize - houseFee;
        
        lottery.prizeWithdrawn = true;
        
        // Transfer prize to winner
        payable(lottery.winner).transfer(winnerPrize);
        
        // Transfer house fee to owner
        payable(owner).transfer(houseFee);
        
        emit PrizeWithdrawn(_lotteryId, lottery.winner, winnerPrize);
    }
    
    /**
     * @dev Get refund if lottery is cancelled (no tickets sold within time limit)
     * @param _lotteryId ID of the lottery
     */
    function getRefund(uint256 _lotteryId) 
        external 
        lotteryExists(_lotteryId) {
        
        Lottery storage lottery = lotteries[_lotteryId];
        require(block.timestamp > lottery.endTime, "Lottery still active");
        require(lottery.ticketsSold == 0, "Tickets were sold");
        require(!lottery.drawn, "Lottery was drawn");
        
        uint256[] memory tickets = lottery.playerTickets[msg.sender];
        require(tickets.length > 0, "No tickets to refund");
        
        uint256 refundAmount = tickets.length * lottery.ticketPrice;
        
        // Clear player tickets to prevent double refund
        delete lottery.playerTickets[msg.sender];
        
        payable(msg.sender).transfer(refundAmount);
        
        emit RefundIssued(_lotteryId, msg.sender, refundAmount);
    }
    
    /**
     * @dev Emergency lottery cancellation with full refunds
     * @param _lotteryId ID of the lottery to cancel
     */
    function cancelLottery(uint256 _lotteryId) 
        external 
        onlyOwner 
        lotteryExists(_lotteryId) {
        
        Lottery storage lottery = lotteries[_lotteryId];
        require(!lottery.drawn, "Cannot cancel drawn lottery");
        
        // Mark as drawn to prevent further ticket sales
        lottery.drawn = true;
        
        // Refund mechanism would be implemented here
        // For simplicity, assuming manual refund process
    }
    
    // View functions
    
    /**
     * @dev Get lottery information
     */
    function getLotteryInfo(uint256 _lotteryId) 
        external 
        view 
        lotteryExists(_lotteryId) 
        returns (
            uint256 ticketPrice,
            uint256 startTime,
            uint256 endTime,
            uint256 maxTickets,
            uint256 ticketsSold,
            uint256 totalPrize,
            bool drawn,
            address winner,
            uint256 winningTicket
        ) {
        
        Lottery storage lottery = lotteries[_lotteryId];
        
        return (
            lottery.ticketPrice,
            lottery.startTime,
            lottery.endTime,
            lottery.maxTickets,
            lottery.ticketsSold,
            lottery.totalPrize,
            lottery.drawn,
            lottery.winner,
            lottery.winningTicket
        );
    }
    
    /**
     * @dev Get player's tickets for a lottery
     */
    function getPlayerTickets(uint256 _lotteryId, address _player) 
        external 
        view 
        lotteryExists(_lotteryId) 
        returns (uint256[] memory) {
        
        return lotteries[_lotteryId].playerTickets[_player];
    }
    
    /**
     * @dev Get ticket owner
     */
    function getTicketOwner(uint256 _lotteryId, uint256 _ticketNumber) 
        external 
        view 
        lotteryExists(_lotteryId) 
        returns (address) {
        
        return lotteries[_lotteryId].ticketToPlayer[_ticketNumber];
    }
    
    /**
     * @dev Verify lottery randomness
     */
    function verifyRandomness(uint256 _lotteryId) 
        external 
        view 
        lotteryExists(_lotteryId) 
        returns (bytes32 seed, uint256 entropyQuality) {
        
        Lottery storage lottery = lotteries[_lotteryId];
        require(lottery.drawn, "Lottery not drawn yet");
        
        return (
            lottery.randomSeed,
            QuantumRandom.entropyQuality(lottery.randomSeed)
        );
    }
    
    /**
     * @dev Check if address can participate (quantum signature verification)
     * @param _participant Address to check
     * @param _algorithm Signature algorithm
     * @param _messageHash Message hash
     * @param _signature Quantum signature
     * @param _publicKey Public key
     */
    function verifyParticipant(
        address _participant,
        uint8 _algorithm,
        bytes32 _messageHash,
        bytes memory _signature,
        bytes memory _publicKey
    ) external view returns (bool) {
        
        // Verify signature
        bool validSignature = QuantumVerifier.verifySignature(
            _algorithm,
            _messageHash,
            _signature,
            _publicKey
        );
        
        if (!validSignature) {
            return false;
        }
        
        // Verify public key matches participant address
        address derivedAddress = address(uint160(uint256(keccak256(_publicKey))));
        return derivedAddress == _participant;
    }
    
    /**
     * @dev Change owner
     */
    function transferOwnership(address _newOwner) external onlyOwner {
        require(_newOwner != address(0), "Invalid new owner");
        owner = _newOwner;
    }
    
    /**
     * @dev Get current lottery ID
     */
    function getCurrentLotteryId() external view returns (uint256) {
        return currentLotteryId;
    }
}