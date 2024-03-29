defmodule Auction.User do
  import Ecto.Changeset
  use Ecto.Schema

  schema "users" do
    field :username, :string
    field :email_address, :string
    field :password, :string, virtual: true
    field :hashed_password, :string
    has_many :bids, Auction.Bid
    timestamps()
  end

  def changeset(user, params \\ %{}) do
    user
    |> cast(params, [:username, :email_address])
    |> validate_required([:username, :email_address, :hashed_password])
    |> validate_length(:username, min: 3)
    |> unique_constraint(:username)
  end

  def changeset_with_password(user, params \\ %{}) do
    user
    |> cast(params, [:password])
    |> validate_required([:password])
    |> validate_length(:password, min: 8)
    |> validate_confirmation(:password, required: true)
    |> hash_password()
    |> changeset(params)
  end

  defp hash_password(%Ecto.Changeset{changes: %{password: password}} = changeset) do
    changeset
    |> put_change(:hashed_password, Auction.Password.hash(password))
  end
  defp hash_password(changeset), do: changeset
end
