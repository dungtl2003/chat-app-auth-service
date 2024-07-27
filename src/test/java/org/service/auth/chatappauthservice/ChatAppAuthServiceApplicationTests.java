package org.service.auth.chatappauthservice;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.json.JsonMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import jakarta.servlet.http.Cookie;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.entity.enums.TokenState;
import org.service.auth.chatappauthservice.entity.enums.TokenType;
import org.service.auth.chatappauthservice.service.AuthTokenService;
import org.service.auth.chatappauthservice.service.UserService;
import org.service.auth.chatappauthservice.utils.UserDTOMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.core.io.ClassPathResource;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.mock.web.MockHttpServletResponse;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders;
import org.springframework.test.web.servlet.result.MockMvcResultMatchers;

import java.io.IOException;
import java.io.InputStream;
import java.util.Arrays;
import java.util.List;
import java.util.Objects;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@AutoConfigureMockMvc
class ChatAppAuthServiceApplicationTests {

	private static final int PORT = Integer.parseInt(System.getenv("PORT"));

	private static final String VERSION = "v1";

	private static final String DOMAIN = "localhost";

	private static final String API_URL;

	private static final String AUTHENTICATE_URL;

	private static final String REFRESH_URL;

	private static final String AUTHORIZATION_URL;

	private static List<User> tempUsers;

	static {
        API_URL = STR."http://\{DOMAIN}:\{PORT}/api/\{VERSION}/auth";
        AUTHENTICATE_URL = STR."\{API_URL}/login";
        AUTHORIZATION_URL = STR."\{API_URL}/authorize";
        REFRESH_URL = STR."\{API_URL}/refresh";
    }

	@Autowired
	private MockMvc mockMvc;

	@Autowired
	private UserService userService;

	@Autowired
	private AuthTokenService authTokenService;

	@BeforeAll
	public static void setup() {
		tempUsers = getSampleUserFromJson();
	}

	public static JsonNode convertStringToJson(String msg) throws JsonProcessingException {
		ObjectMapper mapper = new ObjectMapper();
		return mapper.readTree(msg);
	}

	private static String buildAuthenticateJsonBodyRequest(User user) throws JsonProcessingException {
		ObjectMapper mapper = JsonMapper.builder().findAndAddModules().build();
		ObjectNode root = mapper.createObjectNode();

		ObjectNode metadata = mapper.createObjectNode();
		metadata.put("method", "authenticate_user");

		ObjectNode payload = mapper.createObjectNode();

		ObjectNode userJson = mapper.createObjectNode();
		userJson.put("email", user.getEmail());
		userJson.put("password", user.getPassword());

		payload.set("user", userJson);

		root.set("metadata", metadata);
		root.set("payload", payload);

		return mapper.writeValueAsString(root);
	}

	private static List<User> getSampleUserFromJson() {
		try (InputStream is = new ClassPathResource("user-sample.json").getInputStream()) {
			ObjectMapper mapper = JsonMapper.builder().findAndAddModules().build();
			return Arrays.asList(mapper.readValue(is, User[].class));
		}
		catch (IOException e) {
			throw new RuntimeException(e);
		}
	}

	private String[] createUserRefreshTokensFamily(User user, long minExp, long maxExp, int numberOfTokens) {
		if (minExp > maxExp || minExp < 0) {
			throw new RuntimeException("Invalid min-max value");
		}

		if (numberOfTokens <= 0) {
			throw new RuntimeException("Number of tokens should be greater than 0");
		}

		String[] refreshTokens = new String[numberOfTokens];
		for (int i = 0; i < numberOfTokens; i++) {
			refreshTokens[i] = authTokenService.createRefreshToken(new UserDTOMapper().apply(user),
					minExp + Math.round(Math.random() * (maxExp - minExp)));
		}

		return refreshTokens;
	}

	private void addTempUsers() {
		tempUsers.forEach(user -> userService.add(user.clone()));
	}

	private void removeTempUsers() {
		tempUsers.forEach(user -> userService.deleteUserById(user.getUserId()));
	}

	@Test
	public void testAuthenticateSuccessRequestShouldProvideAccessAndRefreshToken() {
		int size = tempUsers.size();
		User randomUser = tempUsers.get((int) (Math.random() * size));

		try {
			addTempUsers();

			String responseBody = mockMvc
				.perform(MockMvcRequestBuilders.post(AUTHENTICATE_URL)
					.contentType(MediaType.APPLICATION_JSON)
					.content(buildAuthenticateJsonBodyRequest(randomUser)))
				.andExpect(MockMvcResultMatchers.status().isOk())
				.andExpect(MockMvcResultMatchers.header().exists(HttpHeaders.SET_COOKIE))
				.andReturn()
				.getResponse()
				.getContentAsString();

			JsonNode jsonBody = convertStringToJson(responseBody);
			assertNotNull(jsonBody.get("access_token"));
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
		finally {
			removeTempUsers();
		}
	}

	// @Test
	// public void testAccessTokenShouldHaveRightFormat() {
	// int size = tempUsers.size();
	// User randomUser = tempUsers.get((int) (Math.random() * size));
	//
	// try {
	// addTempUsers();
	//
	// MvcResult result = mockMvc
	// .perform(MockMvcRequestBuilders.post(authenticateUrl)
	// .contentType(MediaType.APPLICATION_JSON)
	// .content(buildAuthenticateJsonBodyRequest(randomUser)))
	// .andReturn();
	//
	// String accessToken =
	// result.getResponse().getHeader(HttpHeaders.AUTHORIZATION);
	// assertNotNull(accessToken);
	// assertEquals(accessToken.indexOf("Bearer"), 0);
	// }
	// catch (Exception e) {
	// fail(e.getMessage());
	// }
	// finally {
	// removeTempUsers();
	// }
	// }

	@Test
	public void testAuthenticateNonExistedUserRequestShouldGet404NotFound() {
		int size = tempUsers.size();
		User randomUser = tempUsers.get((int) (Math.random() * size));

		try {
			mockMvc
				.perform(MockMvcRequestBuilders.post(AUTHENTICATE_URL)
					.contentType(MediaType.APPLICATION_JSON)
					.content(buildAuthenticateJsonBodyRequest(randomUser)))
				.andExpect(MockMvcResultMatchers.status().isNotFound());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
	}

	@Test
	public void testAuthenticateInvalidUserRequestShouldGet422UnprocessableEntity() {
		int size = tempUsers.size();
		User randomUser = tempUsers.get((int) (Math.random() * size));

		try {
			addTempUsers();

			randomUser.setPassword("fakepassword"); // change password to invalid one

			mockMvc
				.perform(MockMvcRequestBuilders.post(AUTHENTICATE_URL)
					.contentType(MediaType.APPLICATION_JSON)
					.content(buildAuthenticateJsonBodyRequest(randomUser)))
				.andExpect(MockMvcResultMatchers.status().isUnprocessableEntity());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
		finally {
			removeTempUsers();
		}
	}

	@Test
	public void testAccessAuthenticateApiWithWrongHttpMethodShouldGet405MethodNotAllowed() {
		try {
			mockMvc.perform(MockMvcRequestBuilders.get(AUTHENTICATE_URL))
				.andExpect(MockMvcResultMatchers.status().isMethodNotAllowed());

			mockMvc.perform(MockMvcRequestBuilders.patch(AUTHENTICATE_URL))
				.andExpect(MockMvcResultMatchers.status().isMethodNotAllowed());

			mockMvc.perform(MockMvcRequestBuilders.put(AUTHENTICATE_URL))
				.andExpect(MockMvcResultMatchers.status().isMethodNotAllowed());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
	}

	@Test
	public void testAuthorizeWithoutTokenShouldGet401UnauthorizedWithMissingCredentialMessage() {
		try {
			String body = mockMvc.perform(MockMvcRequestBuilders.get(AUTHORIZATION_URL))
				.andExpect(MockMvcResultMatchers.status().isUnauthorized())
				.andReturn()
				.getResponse()
				.getContentAsString();

			JsonNode bodyJson = convertStringToJson(body);
			assertEquals("Missing credential", bodyJson.get("message").asText());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
	}

	@Test
	public void testAuthorizeWithWrongTokenFormatShouldGet401UnauthorizedWithInvalidFormatMessage() {
		try {
			String body = mockMvc.perform(MockMvcRequestBuilders.get(AUTHORIZATION_URL).header("authorization", "abcd"))
				.andExpect(MockMvcResultMatchers.status().isUnauthorized())
				.andReturn()
				.getResponse()
				.getContentAsString();

			JsonNode bodyJson = convertStringToJson(body);
			assertEquals("Invalid format", bodyJson.get("message").asText());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
	}

	@Test
    public void testAuthorizeWithInvalidTokenShouldGet401UnauthorizedWithInvalidTokenMessage() {
        String fakeToken = "fakeheader.fakepayload.signature";
        try {
            String body = mockMvc
                    .perform(MockMvcRequestBuilders.get(AUTHORIZATION_URL).header("authorization",
                            STR."Bearer \{fakeToken}"))
                    .andExpect(MockMvcResultMatchers.status().isUnauthorized())
                    .andReturn()
                    .getResponse()
                    .getContentAsString();

            JsonNode bodyJson = convertStringToJson(body);
            assertEquals("Invalid token", bodyJson.get("message").asText());
        } catch (Exception e) {
            fail(e.getMessage());
        }
    }

	@Test
    public void testAuthorizeWithExpiredTokenShouldGet401UnauthorizedWithInvalidTokenMessage() {
        String expiredToken = authTokenService.createAccessToken(new UserDTOMapper().apply(tempUsers.getFirst()), 1);
        try {
            String body = mockMvc
                    .perform(MockMvcRequestBuilders.get(AUTHORIZATION_URL).header("authorization",
                            STR."Bearer \{expiredToken}"))
                    .andExpect(MockMvcResultMatchers.status().isUnauthorized())
                    .andReturn()
                    .getResponse()
                    .getContentAsString();

            JsonNode bodyJson = convertStringToJson(body);
            assertEquals("Invalid token", bodyJson.get("message").asText());
        } catch (Exception e) {
            fail(e.getMessage());
        }
    }

	@Test
    public void testAuthorizeWithValidTokenShouldGet200OkWithAuthorizedMessage() {
        String validToken = authTokenService.createAccessToken(new UserDTOMapper().apply(tempUsers.getFirst()));
        try {
            String body = mockMvc
                    .perform(MockMvcRequestBuilders
                            .get(AUTHORIZATION_URL)
                            .header("authorization", STR."Bearer \{validToken}"))
                    .andExpect(MockMvcResultMatchers.status().isOk())
                    .andReturn().getResponse().getContentAsString();

            JsonNode bodyJson = convertStringToJson(body);
            assertEquals("Authorized", bodyJson.get("message").asText());
        } catch (Exception e) {
            fail(e.getMessage());
        }
    }

	@Test
	public void testRefreshWithoutRefreshTokenShouldGet400BadRequestWithMissingRefreshTokenMessage() {
		try {
			String body = mockMvc.perform(MockMvcRequestBuilders.get(REFRESH_URL))
				.andExpect(MockMvcResultMatchers.status().isBadRequest())
				.andReturn()
				.getResponse()
				.getContentAsString();

			JsonNode bodyJson = convertStringToJson(body);
			assertEquals("Missing refresh token", bodyJson.get("message").asText());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
	}

	@Test
	public void testRefreshWithInvalidRefreshTokenShouldGet404NotFoundWithInvalidRefreshTokenMessage() {
		String fakeRefreshToken = "fakeheader.fakepayload.signature";
		Cookie cookie = new Cookie("refresh_token", fakeRefreshToken);
		cookie.setMaxAge(7 * 24 * 60 * 60);

		try {
			String body = mockMvc.perform(MockMvcRequestBuilders.get(REFRESH_URL).cookie(cookie))
				.andExpect(MockMvcResultMatchers.status().isUnauthorized())
				.andReturn()
				.getResponse()
				.getContentAsString();

			JsonNode bodyJson = convertStringToJson(body);
			assertEquals("Invalid refresh token", bodyJson.get("message").asText());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
	}

	@Test
	public void testRefreshWithExpiredRefreshTokenShouldGet401UnauthorizedWithExpiredMessage() {
		String[] refreshTokensFamily = createUserRefreshTokensFamily(tempUsers.getFirst(), 1, 1, 1);
		tempUsers.getFirst().setRefreshTokens(refreshTokensFamily);

		Cookie cookie = new Cookie("refresh_token", refreshTokensFamily[0]);
		cookie.setMaxAge(7 * 24 * 60 * 60);

		try {
			addTempUsers();

			String body = mockMvc.perform(MockMvcRequestBuilders.get(REFRESH_URL).cookie(cookie))
				.andExpect(MockMvcResultMatchers.status().isUnauthorized())
				.andReturn()
				.getResponse()
				.getContentAsString();

			JsonNode bodyJson = convertStringToJson(body);
			assertEquals("Token is expired", bodyJson.get("message").asText());
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
		finally {
			removeTempUsers();
		}
	}

	@Test
	public void testRefreshWithValidRefreshTokenShouldProvideBothTokens() {
		String refreshToken = authTokenService.createRefreshToken(new UserDTOMapper().apply(tempUsers.getFirst()));
		String[] refreshTokens = new String[1];
		refreshTokens[0] = refreshToken;
		tempUsers.getFirst().setRefreshTokens(refreshTokens);

		Cookie cookie = new Cookie("refresh_token", refreshToken);
		cookie.setMaxAge(7 * 24 * 60 * 60);

		try {
			addTempUsers();

			MockHttpServletResponse response = mockMvc.perform(MockMvcRequestBuilders.get(REFRESH_URL).cookie(cookie))
				.andExpect(MockMvcResultMatchers.status().isOk())
				.andExpect(MockMvcResultMatchers.header().exists(HttpHeaders.SET_COOKIE))
				.andReturn()
				.getResponse();

			String newRefreshToken = Objects.requireNonNull(response.getCookie("refresh_token")).getValue();
			assertEquals(authTokenService.checkTokenState(newRefreshToken, TokenType.REFRESH_TOKEN), TokenState.VALID);
			assertNotEquals(refreshToken, newRefreshToken);

			String responseBody = response.getContentAsString();
			JsonNode jsonBody = convertStringToJson(responseBody);
			assertNotNull(jsonBody.get("access_token"));
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
		finally {
			removeTempUsers();
		}
	}

	@Test
	public void testDetectReusedValidRefreshTokenShouldGet401UnauthorizedAndInvalidateAllDevices() {
		String firstRefreshToken = authTokenService.createRefreshToken(new UserDTOMapper().apply(tempUsers.getFirst()));
		String[] refreshTokens = new String[1];
		refreshTokens[0] = firstRefreshToken;
		tempUsers.getFirst().setRefreshTokens(refreshTokens);

		try {
			addTempUsers();

			// normal user sends request with the first token (not used) should be fine
			Cookie firstRTCookie = new Cookie("refresh_token", firstRefreshToken);
			firstRTCookie.setMaxAge(7 * 24 * 60 * 60);
			String secondRefreshToken = Objects
				.requireNonNull(mockMvc.perform(MockMvcRequestBuilders.get(REFRESH_URL).cookie(firstRTCookie))
					.andExpect(MockMvcResultMatchers.status().isOk())
					.andReturn()
					.getResponse()
					.getCookie("refresh_token"))
				.getValue();

			// normal user sends next request with new refresh token should work normally
			Thread.sleep(1000);
			Cookie secondRTCookie = new Cookie("refresh_token", secondRefreshToken);
			secondRTCookie.setMaxAge(7 * 24 * 60 * 60);
			String thirdRefreshToken = Objects
				.requireNonNull(mockMvc.perform(MockMvcRequestBuilders.get(REFRESH_URL).cookie(secondRTCookie))
					.andExpect(MockMvcResultMatchers.status().isOk())
					.andReturn()
					.getResponse()
					.getCookie("refresh_token"))
				.getValue();

			// hacker tries to reuse the second token
			Thread.sleep(1000);
			mockMvc.perform(MockMvcRequestBuilders.get(REFRESH_URL).cookie(secondRTCookie))
				.andExpect(MockMvcResultMatchers.status().isUnauthorized())
				.andReturn();

			// now, no tokens should work
			Thread.sleep(1000);
			Cookie thirdRTCookie = new Cookie("refresh_token", thirdRefreshToken);
			thirdRTCookie.setMaxAge(7 * 24 * 60 * 60);
			mockMvc.perform(MockMvcRequestBuilders.get(REFRESH_URL).cookie(thirdRTCookie))
				.andExpect(MockMvcResultMatchers.status().isUnauthorized())
				.andReturn();

			Thread.sleep(1000);
			mockMvc.perform(MockMvcRequestBuilders.get(REFRESH_URL).cookie(firstRTCookie))
				.andExpect(MockMvcResultMatchers.status().isUnauthorized())
				.andReturn();
		}
		catch (Exception e) {
			fail(e.getMessage());
		}
		finally {
			removeTempUsers();
		}
	}

}
