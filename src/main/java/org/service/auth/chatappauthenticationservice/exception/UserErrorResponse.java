package org.service.auth.chatappauthenticationservice.exception;

public record UserErrorResponse(int status, String message, long timestamp) {
}
