defmodule AuctionWeb.SessionController do
  use AuctionWeb, :controller

  def new(conn, _params), do: render(conn, "new.html")

  def create(conn, %{"user" => %{"username" => username, "password" => password}}) do
    case Auction.get_user_by_username_and_password(username, password) do
      %Auction.User{} = user ->
        conn
        |> put_session(:user_id, user.id)
        |> put_flash(:info, "Welcome, #{user.username}")
        |> redirect(to: Routes.user_path(conn, :show, user))
      _ ->
        conn
        |> put_flash(:error, "That password is incorrect")
        |> render("new.html")
    end
  end

  def delete(conn, _params) do
    conn
    |> clear_session()
    |> configure_session(drop: true)
    |> redirect(to: Routes.item_path(conn, :index))
  end
end
