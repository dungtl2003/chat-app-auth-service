package org.service.auth.chatappauthservice.controller;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.ExampleObject;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import jakarta.servlet.http.HttpServletRequest;
import org.service.auth.chatappauthservice.exception.authorize.InvalidAuthorizationHeaderException;
import org.service.auth.chatappauthservice.exception.authorize.MissingAccessTokenException;
import org.service.auth.chatappauthservice.exception.refresh.MissingRefreshTokenException;
import org.service.auth.chatappauthservice.exception.token.InvalidTokenException;
import org.service.auth.chatappauthservice.exception.user.InvalidUserException;
import org.service.auth.chatappauthservice.exception.user.UserNotFoundException;
import org.service.auth.chatappauthservice.response.AuthenticationResponse;
import org.service.auth.chatappauthservice.response.AuthorizationResponse;
import org.service.auth.chatappauthservice.response.RefreshResponse;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RequestMapping("/default")
public interface AuthApi {

	@Operation(summary = "Logs user into the system",
			responses = { @ApiResponse(responseCode = "200", description = "Success"), },
			requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(required = true,
					content = @Content(examples = @ExampleObject(value = """
							{
							    "payload" : {
							        "user" : {
							            "email": "toan@gmail.com",
							            "password" : "12345678"
							        }
							    }
							}
							"""))))
	@PostMapping(value = "/login", consumes = "application/json")
	ResponseEntity<AuthenticationResponse> authenticate(@RequestBody JsonNode request)
			throws UserNotFoundException, InvalidUserException;

	@Operation(summary = "Check user's access token validation",
			responses = { @ApiResponse(responseCode = "200", description = "Authorized"), },
			security = @SecurityRequirement(name = "authorization"))
	@GetMapping(value = "/authorize", produces = "application/json")
	ResponseEntity<AuthorizationResponse> authorize(@RequestHeader Map<String, String> headers)
			throws MissingAccessTokenException, InvalidAuthorizationHeaderException, InvalidTokenException;

	@Operation(summary = "Refresh user's token",
			responses = { @ApiResponse(responseCode = "200", description = "Success"), },
			security = @SecurityRequirement(name = "refresh_token"))
	@GetMapping(value = "/refresh", produces = "application/json")
	ResponseEntity<RefreshResponse> refresh(HttpServletRequest request)
			throws MissingRefreshTokenException, InvalidTokenException, JsonProcessingException;

}
