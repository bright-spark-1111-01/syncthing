angular.module('syncthing.core')
    .controller('AgentSessionsController', function ($scope, $http, $translate) {
        $scope.sessions = [];
        $scope.selectedSession = null;
        $scope.searchQuery = '';
        $scope.loading = false;
        $scope.error = null;

        // Load all agent sessions
        $scope.loadSessions = function () {
            $scope.loading = true;
            $scope.error = null;

            var url = 'rest/agent/sessions';
            if ($scope.searchQuery) {
                url += '?query=' + encodeURIComponent($scope.searchQuery);
            }

            $http.get(url)
                .then(function (response) {
                    $scope.sessions = response.data || [];
                    $scope.loading = false;
                }, function (error) {
                    console.error('Failed to load agent sessions:', error);
                    $scope.error = 'Failed to load agent sessions';
                    $scope.loading = false;
                });
        };

        // View details of a specific session
        $scope.viewSession = function (session) {
            $scope.selectedSession = session;
        };

        // Close session details view
        $scope.closeSessionDetails = function () {
            $scope.selectedSession = null;
        };

        // Delete a session
        $scope.deleteSession = function (sessionId) {
            if (!confirm($translate.instant('Are you sure you want to delete this session?'))) {
                return;
            }

            $http.delete('rest/agent/sessions/' + sessionId)
                .then(function () {
                    $scope.loadSessions();
                    if ($scope.selectedSession && $scope.selectedSession.id === sessionId) {
                        $scope.closeSessionDetails();
                    }
                }, function (error) {
                    console.error('Failed to delete session:', error);
                    alert('Failed to delete session');
                });
        };

        // Search sessions
        $scope.search = function () {
            $scope.loadSessions();
        };

        // Format timestamp
        $scope.formatTimestamp = function (timestamp) {
            return new Date(timestamp).toLocaleString();
        };

        // Get status badge class
        $scope.getStatusClass = function (status) {
            switch (status) {
                case 'completed':
                    return 'success';
                case 'in_progress':
                    return 'info';
                case 'failed':
                    return 'danger';
                default:
                    return 'default';
            }
        };

        // Initialize: load sessions on controller load
        $scope.loadSessions();
    });
