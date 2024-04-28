package org.service.auth.chatappauthservice.response;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class RefreshResponse {

	@JsonProperty("access_token")
	private String accessToken;

}
