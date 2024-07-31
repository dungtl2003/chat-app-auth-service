package org.service.auth.chatappauthservice.exception.refresh;

import org.service.auth.chatappauthservice.exception.ErrorResponse;
import org.service.auth.chatappauthservice.exception.token.ExpiredTokenException;
import org.service.auth.chatappauthservice.exception.token.InvalidTokenException;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;

@ControllerAdvice
public class RefreshExceptionHandler {

	@ExceptionHandler(value = { MissingRefreshTokenException.class, ReusedRefreshTokenException.class })
	public ResponseEntity<ErrorResponse> handleException(RuntimeException exception) {
		ErrorResponse error = new ErrorResponse(HttpStatus.UNAUTHORIZED.value(), exception.getMessage(),
				System.currentTimeMillis());

		return new ResponseEntity<>(error, HttpStatus.UNAUTHORIZED);
	}

}
