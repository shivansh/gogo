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

runtime:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end runtime


	.globl main
	.ent main
main:
	li	$2, 4
	la	$4, nStr
	syscall
	li	$2, 5
	syscall
	move	$3, $2
	move	$5, $3		# ans -> $5
	# Store dirty variables back into memory
	sw	$3, x
	sw	$5, ans
	beq	$3, 0, exit

	lw	$3, x		# x -> $3
	beq	$3, 1, exit

	li	$3, 1		# start -> $3
	sw	$3, start	# spilled start, freed $3
	lw	$3, x		# x -> $3
	move	$5, $3		# end -> $5
	# Store dirty variables back into memory
	sw	$5, end

while:
	lw	$3, start	# start -> $3
	lw	$5, end		# end -> $5
	add	$6, $3, $5
	srl	$6, $6, 1
	mul	$3, $6, $6
	# x is a perfect square
	lw	$5, x		# x -> $5
	# Store dirty variables back into memory
	sw	$3, temp
	sw	$6, mid
	beq	$3, $5, perfectSquare

	lw	$3, temp	# temp -> $3
	lw	$5, x		# x -> $5
	blt	$3, $5, ifBranch

	lw	$3, mid		# mid -> $3
	sub	$5, $3, 1
	lw	$3, start	# start -> $3
	# Store dirty variables back into memory
	sw	$5, end
	ble	$3, $5, while

	j	exit

ifBranch:
	lw	$3, mid		# mid -> $3
	addi	$5, $3, 1
	move	$6, $3		# ans -> $6
	sw	$6, ans		# spilled ans, freed $6
	lw	$6, end		# end -> $6
	# Store dirty variables back into memory
	sw	$5, start
	ble	$5, $6, while

	j	exit

perfectSquare:
	lw	$3, mid		# mid -> $3
	move	$5, $3		# ans -> $5
	# Store dirty variables back into memory
	sw	$5, ans

exit:
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 1
	lw	$3, ans		# ans -> $3
	move	$4, $3
	syscall
	li	$2, 10
	syscall
	.end main
