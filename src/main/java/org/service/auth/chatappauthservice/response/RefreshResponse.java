package org.service.auth.chatappauthservice.response;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class RefreshResponse {

	private String accessToken;

}
