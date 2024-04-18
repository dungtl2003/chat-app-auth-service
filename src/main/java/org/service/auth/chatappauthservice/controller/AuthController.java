package org.service.auth.chatappauthservice.controller;

import com.fasterxml.jackson.databind.JsonNode;
import lombok.AllArgsConstructor;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.service.auth.chatappauthservice.exception.client.InvalidUserException;
import org.service.auth.chatappauthservice.exception.client.UserNotFoundException;
import org.service.auth.chatappauthservice.response.AuthenticationResponse;
import org.service.auth.chatappauthservice.response.AuthorizationResponse;
import org.service.auth.chatappauthservice.response.RefreshResponse;
import org.service.auth.chatappauthservice.service.AuthService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@AllArgsConstructor
@RequestMapping("/api/v1/auth")
public class AuthController {

	private static final Logger logger = LogManager.getLogger(AuthController.class);

	private final AuthService authService;

	@PostMapping("/login")
    public ResponseEntity<AuthenticationResponse> authenticate(@RequestBody JsonNode request) throws UserNotFoundException, InvalidUserException {
        logger.info(STR."Received request: \{request}");
        return authService.login(request);
    }

	@GetMapping("/authorize")
    public ResponseEntity<AuthorizationResponse> authorize(@RequestHeader Map<String, String> headers) {
        logger.info(STR."Received headers: \{headers}");
        return authService.authorize(headers);
    }

	// TODO: do later
	@GetMapping("/refresh")
    public ResponseEntity<RefreshResponse> refresh(@CookieValue("jwt") String jwtToken) {
        logger.info(STR."Received cookie: \{jwtToken}");
        return authService.refresh(null);
    }

}
