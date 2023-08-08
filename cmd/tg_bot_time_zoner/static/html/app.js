// Function to parse URL data
function parseUrlData() {
	const fragment = window.location.hash.substring(1);
	const params = new URLSearchParams(fragment);
	const userParam = params.get('tgWebAppData');
	const decodedUserParam = decodeURIComponent(userParam);
	const userJson = decodedUserParam.substring(decodedUserParam.indexOf('{'), decodedUserParam.lastIndexOf('}') + 1);
	const user = JSON.parse(decodeURIComponent(userJson));
	return user;
}

// Function to load the profile page
function loadProfilePage(username) {
	fetch(`/api/profile?username=${encodeURIComponent(username)}`)
		.then(response => response.json())
		.then(data => {
			// Display the user's profile data
			document.getElementById('firstName1').value = data.first_name;
			document.getElementById('lastName1').value = data.last_name;
			document.getElementById('username1').value = data.username;
			document.getElementById('description').value = data.description;
			document.getElementById('keywords').value = data.keywords;
		})
		.catch(error => {
			console.error('There has been a problem with your fetch operation:', error);
		});
}
// Function to show the login page
function showLoginPage() {
	document.getElementById('userRegisterDiv').style.display = 'block';
	document.getElementById('loginDiv').style.display = 'block';
}

// Function to login the user
function loginUser(user) {
	fetch('/api/login', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({
			username: user.username
		})}).then(response => response.json()).then(data => {
		if (data.success) {
			// User is authenticated, load the profile page
			loadProfilePage(user.username);
		} else {
			// User is not authenticated, show the login page
			showLoginPage();
		}
	});
}