// Content script

let ws;
let userName;

// Function to connect to the WebSocket server
function connectToWebSocket() {
    ws = new WebSocket('ws://localhost:8080/ws');

    ws.onopen = function(event) {
        console.log('Connected to WebSocket server');
    };

    ws.onmessage = function(event) {
        const data = JSON.parse(event.data);
        if (typeof data === 'string') {
            // If the data is a string, it's the list of connected users
            updateUsersList(data);
        } else {
            // If the data is an object, it's the winner information
            displayWinner(data);
        }
    };

    // Check if username is saved locally
    userName = localStorage.getItem('userName');
    if (userName) {
        registerName(userName);
    }
}

// Function to register the user's name
function registerName(name) {
    userName = name
    localStorage.setItem('userName', userName);
    ws.send(JSON.stringify(userName));
}

// Function to start the game
function startGame() {
    ws.send(JSON.stringify('start'));
}

// Function to reset the game
function resetGame() {
    ws.send(JSON.stringify('reset'));
}

// Function to update the list of connected users
function updateUsersList(users) {
    const usersList = document.getElementById('usersList');
    usersList.innerHTML = "<h2>Registered Users:</h2>";
    users.forEach(user => {
        const userItem = document.createElement('div');
        userItem.textContent = user;
        usersList.appendChild(userItem);
    });
}

// Function to display the winner information
function displayWinner(winnerInfo) {
    console.log('Winner:', winnerInfo.winner);
    console.log('Time taken:', winnerInfo.time_taken);
}

// Register button click event
document.getElementById('registerButton').onclick = registerName()
// Start button click event
document.getElementById('startButton').addEventListener('click', function() {
    startGame();
});

// Reset button click event
document.getElementById('resetButton').addEventListener('click', function() {
    resetGame();
});

connectToWebSocket();
