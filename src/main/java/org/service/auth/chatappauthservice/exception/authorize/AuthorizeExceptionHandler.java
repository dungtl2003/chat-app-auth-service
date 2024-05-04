package org.service.auth.chatappauthservice.exception.authorize;

import org.service.auth.chatappauthservice.exception.ErrorResponse;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.ResponseStatus;

@ControllerAdvice
public class AuthorizeExceptionHandler {

	@ExceptionHandler(value = { InvalidAuthorizationHeaderException.class, MissingAccessTokenException.class })
	public ResponseEntity<ErrorResponse> handleException(RuntimeException exception) {
		ErrorResponse error = new ErrorResponse(HttpStatus.UNAUTHORIZED.value(), exception.getMessage(),
				System.currentTimeMillis());

		return new ResponseEntity<>(error, HttpStatus.UNAUTHORIZED);
	}

}
