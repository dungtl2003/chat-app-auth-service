package org.service.auth.chatappauthservice.exception;

public record ErrorResponse(int status, String message, long timestamp) {
}
