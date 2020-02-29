package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	labrpc "github.com/ddeka0/mrgo/src/labrpc"
)

const (
	// STATE_FOLLOWER .. Follwer state of raft server
	STATE_FOLLOWER = 0
	// STATE_CANDIDATE .. Follwer state of raft server
	STATE_CANDIDATE = 1
	// STATE_LEADER .. Follwer state of raft server
	STATE_LEADER = 2

	MAX_TIMEOUT = 300
	MIN_TIMEOUT = 150
)

// import "bytes"
// import "../labgob"

// ApplyMsg ...
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

// Raft ...
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	// These are the states added from the Fig2 of raft paper
	CurrentTerm          int
	VotedFor             int
	commitIndex          int
	lastApplied          int
	nextIndex            []int
	matchIndex           []int
	CurrentState         int
	ElectionTimeOutTimer *time.Timer
	Mtx                  sync.Mutex
	VoteCount            int
}

// GetState ...
// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	rf.Mtx.Lock()
	term = rf.CurrentTerm
	isleader = (rf.CurrentState == STATE_LEADER)
	rf.Mtx.Unlock()
	return term, isleader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// RequestVoteArgs ...
// Fill in the RequestVoteArgs and RequestVoteReply structs. Modify Make()
// to create a background goroutine that will kick off leader election periodically
// by sending out RequestVote RPCs when it hasn't heard from another peer for a
// while. This way a peer will learn who is the leader, if there is already a
// leader, or become the leader itself. Implement the RequestVote() RPC handler
// so that servers will vote for one another.
type RequestVoteArgs struct {
	// Your data here (2A, 2B).

	// Arguments:
	// term candidate’s term
	// candidateId candidate requesting vote
	// lastLogIndex index of candidate’s last log entry (§5.4)
	// lastLogTerm term of candidate’s last log entry (§5.4)
	Term         int
	CandidateID  int
	LastLogIndex int
	LastLogTerm  int
}

// RequestVoteReply ...
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).

	// Results:
	// term currentTerm, for candidate to update itself
	// voteGranted true means candidate received vote
	Term        int
	VoteGranted int
}

// AppendEntries ...
func (rf *Raft) AppendEntries(args *RequestVoteArgs, reply *RequestVoteReply) {

}

// RequestVote ...
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
}

//
// the tester calls Kill() when a Raft instance won't
// be needed again. for your convenience, we supply
// code to set rf.dead (without needing a lock),
// and a killed() method to test rf.dead in
// long-running loops. you can also add your own
// code to Kill(). you're not required to do anything
// about this, but it may be convenient (for example)
// to suppress debug output from a Kill()ed instance.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.VoteCount = 0
	// Your initialization code here (2A, 2B, 2C).

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// Server when statrs up, starts a follower
	rf.CurrentState = STATE_FOLLOWER
	rf.ElectionTimeOutTimer = time.NewTimer(time.Millisecond * time.Duration(
		rand.Intn(MAX_TIMEOUT-MIN_TIMEOUT+1)+MIN_TIMEOUT))

	go func() {
		// kick off leader election periodically by sending out RequestVote RPCs
		// when it hasn't heard from another peer for a while
		for {
			<-rf.ElectionTimeOutTimer.C
			// We can start a new Election now
			rf.Mtx.Lock()
			if rf.CurrentState == STATE_FOLLOWER || rf.CurrentState == STATE_CANDIDATE {
				rf.CurrentState = STATE_CANDIDATE
				rf.CurrentTerm++
				rf.VotedFor = rf.me
				rf.VoteCount++
				rf.ElectionTimeOutTimer.Reset(time.Millisecond * time.Duration(
					rand.Intn(MAX_TIMEOUT-MIN_TIMEOUT+1)+MIN_TIMEOUT))

				for serverID := range rf.peers {
					var req = RequestVoteArgs{}
					var rep = RequestVoteReply{}
					req.CandidateID = rf.me
					req.LastLogIndex = 0
					req.LastLogTerm = 0
					req.Term = rf.CurrentTerm
					if serverID != me {
						// Create a New Go Routine to Send And Block for Receive
						go func() {
							ok := rf.sendRequestVote(serverID, &req, &rep)
							if !ok {
								log.Println("Error in Sending RequestVote RPC!")
							} else {
								if rep.Term > rf.CurrentTerm {
									rf.Mtx.Lock()
									rf.CurrentState = STATE_FOLLOWER
									rf.Mtx.Unlock()
								}
							}
						}()
					}
				}
				fmt.Println("Election Timeout !!! Sent a New RequestVote RPC !")
				rf.Mtx.Unlock()
			}
		}
	}()

	go func() {
		for {
			// simulates a incoming message at any time
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(20-10+1)+10))
			//fmt.Println("messages arrived at ", v)

			// once any message is received reset the timer to 4 seconds again
			rf.Mtx.Lock()
			rf.ElectionTimeOutTimer.Reset(time.Millisecond * time.Duration(
				rand.Intn(MAX_TIMEOUT-MIN_TIMEOUT+1)+MIN_TIMEOUT))
			//t = time.NewTimer(time.Second * time.Duration(TimeOutTime))

			rf.Mtx.Unlock()
		}
	}()
	return rf
}
