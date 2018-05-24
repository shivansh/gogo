# else branch

# Test to find square root (floor value) of a number

	.data
nStr:		.asciiz "Enter n: "
x:		.word	0
ans:		.word	0
start:		.word	0
end:		.word	0
mid:		.word	0
temp:		.word	0
str:		.asciiz "Square root of n: "

	.text

	.globl main
	.ent main
main:
	li	$v0, 4
	la	$a0, nStr
	syscall
	li	$v0, 5
	syscall
	move	$3, $v0
	move	$7, $3		# ans -> $7
	# Store variables back into memory
	sw	$3, x
	sw	$7, ans
	beq	$3, 0, exit	# exit -> $0

	lw	$3, x
	# Store variables back into memory
	sw	$3, x
	beq	$3, 1, exit	# exit -> $0

	li	$3, 1		# start -> $3
	lw	$7, x
	move	$15, $7		# end -> $15
	# Store variables back into memory
	sw	$3, start
	sw	$7, x
	sw	$15, end

while:
	lw	$3, start
	lw	$7, end
	add	$30, $3, $7	# mid -> $30
	srl	$30, $30, 1	# mid -> $30
	mul	$15, $30, $30	# temp -> $15
	# x is a perfect square
	lw	$16, x
	# Store variables back into memory
	sw	$3, start
	sw	$7, end
	sw	$15, temp
	sw	$16, x
	sw	$30, mid
	beq	$15, $16, perfectSquare	# perfectSquare -> $0

	lw	$3, temp
	lw	$7, x
	# Store variables back into memory
	sw	$3, temp
	sw	$7, x
	blt	$3, $7, ifBranch		# ifBranch -> $0

	lw	$3, mid
	sub	$7, $3, 1	# end -> $7
	lw	$30, start
	# Store variables back into memory
	sw	$3, mid
	sw	$7, end
	sw	$30, start
	ble	$30, $7, while	# while -> $0

	j	exit

ifBranch:
	lw	$3, mid
	addi	$7, $3, 1	# start -> $7
	move	$30, $3		# ans -> $30
	lw	$15, end
	# Store variables back into memory
	sw	$3, mid
	sw	$7, start
	sw	$15, end
	sw	$30, ans
	ble	$7, $15, while	# while -> $0

	j	exit

perfectSquare:
	lw	$3, mid
	move	$7, $3		# ans -> $7
	# Store variables back into memory
	sw	$3, mid
	sw	$7, ans

exit:
	li	$v0, 4
	la	$a0, str
	syscall
	li	$v0, 1
	lw	$3, ans
	move	$a0, $3
	syscall
	# Store variables back into memory
	sw	$3, ans
	li	$v0, 10
	syscall
	.end main
